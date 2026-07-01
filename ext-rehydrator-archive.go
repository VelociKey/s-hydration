package hydration

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (h *SovereignPurifier) Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	destClean := filepath.Clean(dest)

	for _, f := range r.File {
		parts := strings.Split(f.Name, "/")
		relPath := f.Name
		if len(parts) > 1 {
			relPath = strings.Join(parts[1:], string(os.PathSeparator))
		}
		if relPath == "" {
			continue
		}

		fpath := filepath.Join(dest, relPath)
		targetClean := filepath.Clean(fpath)
		if !strings.HasPrefix(targetClean, destClean+string(os.PathSeparator)) && targetClean != destClean {
			return fmt.Errorf("illegal file path in zip archive (Zip Slip traversal attempt): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fpath, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", fpath, err)
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}
		_, _ = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}

func (h *SovereignPurifier) UnTgz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)

	destClean := filepath.Clean(dest)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		relPath := header.Name
		if strings.HasPrefix(relPath, "package/") {
			relPath = strings.TrimPrefix(relPath, "package/")
		}
		if relPath == "" {
			continue
		}
		target := filepath.Join(dest, relPath)
		targetClean := filepath.Clean(target)
		if !strings.HasPrefix(targetClean, destClean+string(os.PathSeparator)) && targetClean != destClean {
			return fmt.Errorf("illegal file path in tar archive (Zip Slip traversal attempt): %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", target, err)
			}
			outFile, err := os.Create(target)
			if err != nil {
				continue
			}
			_, _ = io.Copy(outFile, tr)
			outFile.Close()
		}
	}
	return nil
}

func extractNPMTarballStream(tarballPath, outputDir, pkgName string) error {
	file, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var foundBinary string

	targetName := pkgName
	if parts := strings.Split(targetName, "/"); len(parts) > 0 {
		targetName = parts[len(parts)-1]
	}
	targetName = strings.ReplaceAll(targetName, "-cli", "")

	destClean := filepath.Clean(outputDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := header.Name
		if header.Typeflag == tar.TypeReg {
			isExe := strings.HasSuffix(name, ".exe") || strings.Contains(name, "/bin/") || strings.HasSuffix(name, targetName)
			if isExe {
				cleanName := filepath.Base(name)
				if !strings.HasSuffix(cleanName, ".exe") {
					cleanName += ".exe"
				}
				targetPath := filepath.Join(outputDir, cleanName)
				targetClean := filepath.Clean(targetPath)
				if !strings.HasPrefix(targetClean, destClean+string(os.PathSeparator)) && targetClean != destClean {
					return fmt.Errorf("illegal path in NPM tarball (Zip Slip traversal attempt): %s", header.Name)
				}

				if err := os.MkdirAll(outputDir, 0755); err != nil {
					return err
				}
				out, err := os.Create(targetPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(out, tr); err != nil {
					out.Close()
					return err
				}
				out.Close()
				foundBinary = targetPath
			}
		}
	}
	if foundBinary == "" {
		return fmt.Errorf("no native executable found in npm tarball stream")
	}
	return nil
}

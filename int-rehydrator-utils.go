package hydration

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) error {
	sClean := filepath.Clean(src)
	dClean := filepath.Clean(dst)
	if sClean == dClean {
		return nil
	}

	if _, err := os.Stat(dst); err == nil {
		oldDst := dst + ".old"
		_ = os.Remove(oldDst)
		if errRename := os.Rename(dst, oldDst); errRename == nil {
			defer os.Remove(oldDst)
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		return copyFile(path, targetPath)
	})
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, ".gitroot")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf(".gitroot not found")
		}
		current = parent
	}
}

func backupWorkspace(wsPath string) (string, error) {
	wsName := filepath.Base(wsPath)
	backupRoot := filepath.Join(wsPath, "..", "s-hydrationcache", "c0990-ephemeral-scratch", "backups", wsName)
	_ = os.RemoveAll(backupRoot)
	err := os.MkdirAll(backupRoot, 0755)
	if err != nil {
		return "", err
	}

	srcPath := wsPath
	successSrc := filepath.Join(wsPath, "..", "s-hydrationcache", "c0990-ephemeral-scratch", "last_successful_build", wsName)
	if _, err := os.Stat(successSrc); err == nil {
		srcPath = successSrc
	}

	err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "c0990-ephemeral-scratch" || name == "build-caches" || name == "s-forge" || name == "96000-internal-executables" || name == "97000-internal-toolchains" || name == "99000-internal-actors" {
				return filepath.SkipDir
			}
			return nil
		}
		
		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".mod" || ext == ".sum" || ext == ".harness" || ext == ".wag" || ext == ".webnf" || ext == ".yaml" || ext == ".yml" {
			rel, err := filepath.Rel(srcPath, path)
			if err != nil {
				return nil
			}
			dest := filepath.Join(backupRoot, rel)
			_ = os.MkdirAll(filepath.Dir(dest), 0755)
			_ = copyFile(path, dest)
		}
		return nil
	})
	return backupRoot, err
}

func restoreWorkspace(wsPath string, backupRoot string) error {
	slog.Info("Restoring workspace from backup due to compilation/test failure", "workspace", filepath.Base(wsPath), "backup", backupRoot)
	err := filepath.Walk(backupRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(backupRoot, path)
		if err != nil {
			return nil
		}
		target := filepath.Join(wsPath, rel)
		_ = os.Remove(target)
		_ = os.MkdirAll(filepath.Dir(target), 0755)
		_ = copyFile(path, target)
		return nil
	})
	_ = os.RemoveAll(backupRoot)
	return err
}

func cleanLocalWorkstationExecutables(projectRoot string) {
	hydrationDir := filepath.Join(projectRoot, "00flow", "s-hydration")
	
	_ = os.Remove(filepath.Join(hydrationDir, "s-hydration.exe"))
	_ = os.Remove(filepath.Join(hydrationDir, "int-rehydrator.exe"))
}

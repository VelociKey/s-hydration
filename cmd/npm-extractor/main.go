package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sov.fleet/blake3"
)

type PackageMetadata struct {
	Dist struct {
		Tarball string `json:"tarball"`
	} `json:"dist"`
	Version string `json:"version"`
}

func main() {
	pkgName := flag.String("package", "", "NPM package name (e.g. @google/jules)")
	version := flag.String("version", "latest", "Package version")
	outDir := flag.String("out", "", "Output directory for extracted executable")
	usePodman := flag.Bool("podman", false, "Use network-isolated Podman sandbox for installation")
	flag.Parse()

	if *pkgName == "" {
		fmt.Println("Error: -package is required")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Resolve package tarball URL from registry
	fmt.Printf("Resolving package %s (%s)...\n", *pkgName, *version)
	tarballURL, resolvedVersion, err := resolveTarball(client, *pkgName, *version)
	if err != nil {
		fmt.Printf("Failed to resolve package: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Resolved version: %s\nTarball: %s\n", resolvedVersion, tarballURL)

	scratchDir := `C:\aCogSpaceSeed\00flow\s-hydrationcache\c0990-ephemeral-scratch\npm-sandbox`
	if err := os.MkdirAll(scratchDir, 0755); err != nil {
		fmt.Printf("Failed to create scratch dir: %v\n", err)
		os.Exit(1)
	}

	tarballPath := filepath.Join(scratchDir, "package.tgz")
	fmt.Println("Downloading package tarball...")
	if err := downloadFile(client, tarballURL, tarballPath); err != nil {
		fmt.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}

	var binaryPath string
	if *usePodman {
		fmt.Println("Running containerized installation via Podman...")
		binaryPath, err = runPodmanSandbox(*pkgName, tarballPath, scratchDir)
		if err != nil {
			fmt.Printf("Podman sandbox installation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Extracting tarball in-memory...")
		binaryPath, err = extractTarballBinaries(tarballPath, scratchDir)
		if err != nil {
			fmt.Printf("Tarball extraction failed: %v\n", err)
			os.Exit(1)
		}
	}

	if binaryPath == "" {
		fmt.Println("Error: No executable binary found in package output")
		os.Exit(1)
	}

	// 2. Define output destination in s-forge
	targetName := sanitizePackageName(*pkgName)
	if *outDir == "" {
		*outDir = filepath.Join(`C:\aCogSpaceSeed\00flow\s-forge\94000-external-actors`, targetName)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Printf("Failed to create destination directory: %v\n", err)
		os.Exit(1)
	}

	destPath := filepath.Join(*outDir, targetName+".exe")
	if err := copyFile(binaryPath, destPath); err != nil {
		fmt.Printf("Failed to copy binary to destination: %v\n", err)
		os.Exit(1)
	}

	// 3. Compute Blake3 hash and update SBOM
	hash, err := computeBlake3(destPath)
	if err != nil {
		fmt.Printf("Failed to compute Blake3: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully hydrated: %s\nDestination: %s\nBlake3 Seal: %s\n", targetName, destPath, hash)

	if err := updateSBOM(targetName, resolvedVersion, hash); err != nil {
		fmt.Printf("Failed to update SBOM: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("SBOM successfully updated.")
}

func resolveTarball(client *http.Client, pkg, version string) (string, string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", pkg, version)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	var meta PackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", "", err
	}
	return meta.Dist.Tarball, meta.Version, nil
}

func downloadFile(client *http.Client, url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarballBinaries(tarballPath, outputDir string) (string, error) {
	file, err := os.Open(tarballPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var foundBinary string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for executable files or files in bin directories
		name := header.Name
		if header.Typeflag == tar.TypeReg {
			isExe := strings.HasSuffix(name, ".exe") || strings.Contains(name, "/bin/") || strings.HasSuffix(name, "jules")
			if isExe {
				cleanName := filepath.Base(name)
				if !strings.HasSuffix(cleanName, ".exe") {
					cleanName += ".exe" // Standardize executable suffix
				}
				targetPath := filepath.Join(outputDir, cleanName)
				out, err := os.Create(targetPath)
				if err != nil {
					return "", err
				}
				if _, err := io.Copy(out, tr); err != nil {
					out.Close()
					return "", err
				}
				out.Close()
				foundBinary = targetPath
			}
		}
	}
	return foundBinary, nil
}

func sanitizePackageName(pkg string) string {
	parts := strings.Split(pkg, "/")
	name := parts[len(parts)-1]
	return strings.ReplaceAll(name, "-cli", "")
}

func copyFile(src, dst string) error {
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

func computeBlake3(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := blake3.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("blake3:%x", hasher.Sum(nil)), nil
}

func updateSBOM(name, version, hash string) error {
	sbomPath := `C:\aCogSpaceSeed\00flow\s-forge\90100-rehydration-seed\sbom.external_artifact.webnf`
	content, err := os.ReadFile(sbomPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, name+"|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 8 {
				parts[1] = version
				parts[3] = hash
				parts[6] = time.Now().Format("2006-01-02T15:04:05-07:00")
				lines[i] = strings.Join(parts, "|")
				found = true
				break
			}
		}
	}

	if !found {
		// Append if not exists
		newLine := fmt.Sprintf("%s|%s||%s|0|0|%s|ACTOR||SOURCE", name, version, hash, time.Now().Format("2006-01-02T15:04:05-07:00"))
		lines = append(lines, newLine)
	}

	return os.WriteFile(sbomPath, []byte(strings.Join(lines, "\n")), 0644)
}

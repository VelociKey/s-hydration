package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runPodmanSandbox(pkg, tarballPath, scratchDir string) (string, error) {
	containerfilePath := filepath.Join(scratchDir, "Containerfile")
	containerfileContent := `FROM node:20-alpine
WORKDIR /app
COPY package.tgz /app/package.tgz
RUN npm install -g /app/package.tgz --unsafe-perm
`

	if err := os.WriteFile(containerfilePath, []byte(containerfileContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write Containerfile: %w", err)
	}

	// 1. Build the Podman image
	fmt.Println("Building Podman sandbox image...")
	cmdBuild := exec.Command("podman", "build", "-t", "npm-sandbox-image", "-f", "Containerfile", ".")
	cmdBuild.Dir = scratchDir
	cmdBuild.Stdout = os.Stdout
	cmdBuild.Stderr = os.Stderr
	if err := cmdBuild.Run(); err != nil {
		return "", fmt.Errorf("failed to build Podman image: %w", err)
	}

	// 2. Run the image with network isolation to extract binary
	fmt.Println("Running Podman container with --network none to extract executable...")
	targetName := sanitizePackageName(pkg)
	cmdRun := exec.Command("podman", "run", "--rm", "--network", "none",
		"-v", scratchDir+":/output",
		"npm-sandbox-image",
		"sh", "-c", fmt.Sprintf("cp /usr/local/bin/%s* /output/%s.exe || cp /usr/local/lib/node_modules/%s/bin/* /output/%s.exe || cp /usr/local/bin/jules* /output/jules.exe || true", targetName, targetName, targetName, targetName))
	
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr
	if err := cmdRun.Run(); err != nil {
		return "", fmt.Errorf("failed to run Podman container: %w", err)
	}

	expectedBinary := filepath.Join(scratchDir, targetName+".exe")
	if _, err := os.Stat(expectedBinary); err == nil {
		return expectedBinary, nil
	}

	// Try fallback jules.exe
	fallbackBinary := filepath.Join(scratchDir, "jules.exe")
	if _, err := os.Stat(fallbackBinary); err == nil {
		return fallbackBinary, nil
	}

	return "", fmt.Errorf("no executable found in sandbox output directory")
}

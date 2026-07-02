package hydration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploymentPipeline(t *testing.T) {
	// Create temporary workspace config directory
	tmpDir, err := os.MkdirTemp("", "deploy-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, "c0100-configuration-registry", "110-workspace-configs")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write mock deployment.webnf config
	configContent := `
gcp_project_id: "test-project-123"
staging_bucket: "gs://test-staging-bucket"
artifact_registry_repo: "test-repo"
artifact_registry_base: "us-east1-docker.pkg.dev"
`
	configPath := filepath.Join(configDir, "deployment.webnf")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock config file: %v", err)
	}

	// 1. Verify config parsing
	cfg, err := loadDeploymentConfig(tmpDir)
	if err != nil {
		t.Fatalf("loadDeploymentConfig returned error: %v", err)
	}
	if cfg.ProjectID != "test-project-123" {
		t.Errorf("expected ProjectID 'test-project-123', got %q", cfg.ProjectID)
	}
	if cfg.StagingBucket != "gs://test-staging-bucket" {
		t.Errorf("expected StagingBucket 'gs://test-staging-bucket', got %q", cfg.StagingBucket)
	}
	if cfg.RegistryRepo != "test-repo" {
		t.Errorf("expected RegistryRepo 'test-repo', got %q", cfg.RegistryRepo)
	}
	if cfg.RegistryBase != "us-east1-docker.pkg.dev" {
		t.Errorf("expected RegistryBase 'us-east1-docker.pkg.dev', got %q", cfg.RegistryBase)
	}

	// 2. Verify deploy executor run
	dummyWasmPath := filepath.Join(tmpDir, "s-emulator-pubsub.wasm")
	err = os.WriteFile(dummyWasmPath, []byte("wasm-dummy-payload"), 0644)
	if err != nil {
		t.Fatalf("failed to write dummy wasm file: %v", err)
	}

	outputs := []StagedOutput{
		{LocalPath: dummyWasmPath, GlobalPath: dummyWasmPath},
	}

	err = DeployWasmArtifacts(context.Background(), tmpDir, outputs)
	if err != nil {
		t.Fatalf("DeployWasmArtifacts returned unexpected error: %v", err)
	}
}

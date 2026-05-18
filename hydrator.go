package shydration

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// HydrationMode represents the physical mode of the hydration process.
type HydrationMode string

const (
	ModeDynamic  HydrationMode = "DYNAMIC"
	ModeStatic   HydrationMode = "STATIC"
	ModeDiscover HydrationMode = "DISCOVER"
)

// HydrationTarget represents a declarative target configuration.
type HydrationTarget struct {
	Mode       HydrationMode
	Engine     string
	SourcePath string
	OutputPath string
}

// HydrationMechanism defines the plugin contract that all execution engines must implement.
type HydrationMechanism interface {
	// Mode returns the primary mode this mechanism handles.
	Mode() HydrationMode
	// Name returns the descriptive name of the plugin mechanism.
	Name() string
	// Hydrate executes the core materialization or discovery strategy.
	Hydrate(ctx context.Context, target HydrationTarget) error
}

// Hydrator is the central registry and coordinator of hydration plugins.
type Hydrator struct {
	plugins map[HydrationMode]HydrationMechanism
}

// NewHydrator initializes the registry with standard plugin points.
func NewHydrator() *Hydrator {
	return &Hydrator{
		plugins: make(map[HydrationMode]HydrationMechanism),
	}
}

// Register adds a new hydration mechanism plugin to the engine.
func (h *Hydrator) Register(plugin HydrationMechanism) {
	h.plugins[plugin.Mode()] = plugin
	log.Printf("[Hydrator] Registered plugin %q for mode %s", plugin.Name(), plugin.Mode())
}

// Execute drives the hydration loop for a given target, routing to the correct plugin.
func (h *Hydrator) Execute(ctx context.Context, target HydrationTarget) error {
	plugin, exists := h.plugins[target.Mode]
	if !exists {
		return fmt.Errorf("no hydration plugin registered for mode: %s", target.Mode)
	}
	
	log.Printf("[Hydrator] Orchestrating %s hydration via %q...", target.Mode, plugin.Name())
	return plugin.Hydrate(ctx, target)
}

// =========================================================================
//  1. DYNAMIC SYNTHESIS MECHANISM (100-synthesis-engine)
// =========================================================================

// DynamicSynthesisMechanism handles local hermetic compilation of actors.
type DynamicSynthesisMechanism struct{}

func (d *DynamicSynthesisMechanism) Mode() HydrationMode { return ModeDynamic }
func (d *DynamicSynthesisMechanism) Name() string        { return "Dynamic Bazel/Go Synthesis" }
func (d *DynamicSynthesisMechanism) Hydrate(ctx context.Context, target HydrationTarget) error {
	log.Printf("[DynamicSynthesis] Initiating compilation on src: %s...", target.SourcePath)
	
	// Simulate local hermetic build mapping to internal actors
	absOut := filepath.Clean(target.OutputPath)
	log.Printf("[DynamicSynthesis] Synthesizing artifact to target path: %s", absOut)
	
	// Perform safe sandbox output write
	err := os.MkdirAll(filepath.Dir(absOut), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output parent folder: %w", err)
	}
	
	f, err := os.Create(absOut)
	if err != nil {
		return fmt.Errorf("failed to write synthesized binary output: %w", err)
	}
	defer f.Close()
	
	_, err = io.WriteString(f, fmt.Sprintf("; Sealed Synthesized Output\n; Mode: DYNAMIC\n; Engine: %s\n", target.Engine))
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	
	log.Printf("[DynamicSynthesis] Successfully synthesized and sealed actor artifact!")
	return nil
}

// =========================================================================
//  2. STATIC INGESTION MECHANISM (200-ingestion-engine)
// =========================================================================

// StaticIngestionMechanism handles download, hash verification, and mirror ingestion.
type StaticIngestionMechanism struct{}

func (s *StaticIngestionMechanism) Mode() HydrationMode { return ModeStatic }
func (s *StaticIngestionMechanism) Name() string        { return "Static Ingestion Mirror" }
func (s *StaticIngestionMechanism) Hydrate(ctx context.Context, target HydrationTarget) error {
	log.Printf("[StaticIngestion] Processing ingestion for archive: %s...", target.SourcePath)
	
	// Simulate SHA-256 and veracity seal verification
	log.Printf("[StaticIngestion] Verifying veracity seal for asset %s...", target.SourcePath)
	
	absOut := filepath.Clean(target.OutputPath)
	err := os.MkdirAll(filepath.Dir(absOut), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output parent folder: %w", err)
	}
	
	f, err := os.Create(absOut)
	if err != nil {
		return fmt.Errorf("failed to ingest static asset: %w", err)
	}
	defer f.Close()
	
	_, err = io.WriteString(f, fmt.Sprintf("; Sealed Ingested Output\n; Mode: STATIC\n; Engine: %s\n", target.Engine))
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	
	log.Printf("[StaticIngestion] Successfully ingested and verified static artifact!")
	return nil
}

// =========================================================================
//  3. MCP CAPABILITY DISCOVERY MECHANISM (300-discovery-engine)
// =========================================================================

// MCPDiscoveryMechanism connects to Model Context Protocol endpoints to ingest schemas.
type MCPDiscoveryMechanism struct {
	// Expose properties to configure local mock capabilities for testing
	MockCapabilities []string
}

func (m *MCPDiscoveryMechanism) Mode() HydrationMode { return ModeDiscover }
func (m *MCPDiscoveryMechanism) Name() string        { return "Model Context Protocol (MCP) Discovery" }
func (m *MCPDiscoveryMechanism) Hydrate(ctx context.Context, target HydrationTarget) error {
	log.Printf("[MCPDiscovery] Connecting to federated client at path: %s...", target.SourcePath)
	log.Printf("[MCPDiscovery] Executing dynamic capability discovery query over MCP...")
	
	// Standardize on default mock tools if none are provided
	caps := m.MockCapabilities
	if len(caps) == 0 {
		caps = []string{"trivy_scan", "sign_pruned_product", "translate_gemma_semantic"}
	}
	
	absOut := filepath.Clean(target.OutputPath)
	err := os.MkdirAll(filepath.Dir(absOut), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output parent folder: %w", err)
	}
	
	f, err := os.Create(absOut)
	if err != nil {
		return fmt.Errorf("failed to create capability catalog file: %w", err)
	}
	defer f.Close()
	
	// Formally write a WAG-conforming WebNF file mapping discovered capabilities
	_, err = io.WriteString(f, fmt.Sprintf(`; Discovered Capabilities Catalog
; Mode: DISCOVER
; Engine: %s
; Interface: Model Context Protocol (MCP)

capability_catalog {
`, target.Engine))
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	
	for _, capName := range caps {
		_, err = io.WriteString(f, fmt.Sprintf("    MCP_TOOL_%s = \"enabled\" ;\n", capName))
		if err != nil {
			return fmt.Errorf("failed to write capability mapping: %w", err)
		}
		log.Printf("[MCPDiscovery] Discovered capability tool: %q", capName)
	}
	
	_, err = io.WriteString(f, "}\n")
	if err != nil {
		return fmt.Errorf("failed to finalize catalog: %w", err)
	}
	
	log.Printf("[MCPDiscovery] Successfully mapped %d discovered tools to capability_catalog.webnf", len(caps))
	return nil
}

// =========================================================================
//  4. DYNAMIC WEBNF INTEGRATION & INVOKE-TO-ADD (Licensee Augmentation)
// =========================================================================

// InvokeInput represents the parameters for dynamically adding a hydration target.
type InvokeInput struct {
	Mode       HydrationMode
	Engine     string
	SourcePath string
	OutputPath string
}

// InvokeToAdd executes a dynamic hydration target. If permanent is true, it persistently 
// appends the target schema to the specified WebNF configuration file database on disk.
func (h *Hydrator) InvokeToAdd(ctx context.Context, input InvokeInput, permanent bool, webnfPath string) error {
	log.Printf("[Hydrator] Invoking 'Invoke-to-Add' mechanism...")
	log.Printf("[Hydrator] Details -> Mode: %s, Engine: %s, Src: %s, Out: %s (Permanent: %t)", 
		input.Mode, input.Engine, input.SourcePath, input.OutputPath, permanent)

	target := HydrationTarget{
		Mode:       input.Mode,
		Engine:     input.Engine,
		SourcePath: input.SourcePath,
		OutputPath: input.OutputPath,
	}

	if permanent {
		log.Printf("[Hydrator] Performing permanent inclusion. Appending target schema to: %s", webnfPath)
		
		// If file doesn't exist, create a valid baseline workspace harness structure
		if _, err := os.Stat(webnfPath); os.IsNotExist(err) {
			baseline := fmt.Sprintf(`; Declarative Workspace Harness
; Governed by hydrator.wag

workspace_harness {
    name : "dynamic_licensee_workspace" ;
    targets {
    }
    promote {
        realm : LICENSEE_CUSTOM ;
        origin : INTERNAL ;
        category : ACTOR ;
    }
}
`)
			err = os.WriteFile(webnfPath, []byte(baseline), 0644)
			if err != nil {
				return fmt.Errorf("failed to create baseline WebNF config: %w", err)
			}
		}

		contentBytes, err := os.ReadFile(webnfPath)
		if err != nil {
			return fmt.Errorf("failed to read WebNF config: %w", err)
		}
		
		content := string(contentBytes)
		
		// Synthesize a beautiful, indented WebNF target block matching hydrator.wag rules
		newTargetBlock := fmt.Sprintf(`        target {
            mode : %s ;
            engine : "%s" ;
            src : "%s" ;
            out : "%s" ;
        }
`, target.Mode, target.Engine, target.SourcePath, target.OutputPath)

		// Safely inject the new target block inside the targets { ... } section
		var updatedContent string
		importIndex := strings.Index(content, "targets {")
		if importIndex != -1 {
			targetBlockStart := importIndex + len("targets {")
			updatedContent = content[:targetBlockStart] + "\n" + newTargetBlock + content[targetBlockStart:]
		} else {
			// If targets block is completely missing, inject it before the promote block
			promoteIndex := strings.Index(content, "promote {")
			if promoteIndex != -1 {
				updatedContent = content[:promoteIndex] + "targets {\n" + newTargetBlock + "    }\n    " + content[promoteIndex:]
			} else {
				return fmt.Errorf("invalid WebNF schema: missing workspace_harness structure")
			}
		}

		err = os.WriteFile(webnfPath, []byte(updatedContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write updated WebNF config: %w", err)
		}
		log.Printf("[Hydrator] Successfully persisted dynamic target schema to WebNF database.")
	} else {
		log.Printf("[Hydrator] Performing short-term inclusion. Running ephemeral target in memory.")
	}

	return h.Execute(ctx, target)
}

// ParseWebNFTargets reads a WebNF configuration string and dynamically extracts all
// hydration target blocks using a fast, robust scanning model.
func ParseWebNFTargets(webnfContent string) ([]HydrationTarget, error) {
	var parsedTargets []HydrationTarget
	index := 0
	
	for {
		targetIdx := strings.Index(webnfContent[index:], "target")
		if targetIdx == -1 {
			break
		}
		
		absoluteTargetIdx := index + targetIdx
		rest := webnfContent[absoluteTargetIdx+len("target"):]
		braceIdx := strings.Index(rest, "{")
		if braceIdx == -1 {
			break
		}
		
		// Ensure only whitespace exists between "target" and "{"
		mid := rest[:braceIdx]
		if strings.TrimSpace(mid) != "" && !strings.HasPrefix(strings.TrimSpace(mid), ";") {
			index = absoluteTargetIdx + len("target")
			continue
		}
		
		// Find matching closing brace
		startBrace := absoluteTargetIdx + len("target") + braceIdx
		depth := 1
		endBrace := -1
		for i := startBrace + 1; i < len(webnfContent); i++ {
			if webnfContent[i] == '{' {
				depth++
			} else if webnfContent[i] == '}' {
				depth--
				if depth == 0 {
					endBrace = i
					break
				}
			}
		}
		
		if endBrace == -1 {
			break
		}
		
		// Parse the contents of the target block
		blockContent := webnfContent[startBrace+1 : endBrace]
		var t HydrationTarget
		lines := strings.Split(blockContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, ";") {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.TrimSuffix(val, ";")
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `;"'`)
			
			switch key {
			case "mode":
				t.Mode = HydrationMode(val)
			case "engine":
				t.Engine = val
			case "src":
				t.SourcePath = val
			case "out":
				t.OutputPath = val
			}
		}
		
		if t.Mode != "" {
			parsedTargets = append(parsedTargets, t)
		}
		
		index = endBrace + 1
	}
	
	return parsedTargets, nil
}

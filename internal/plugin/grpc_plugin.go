package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivannovak/glide-plugin-go/pkg/version"
	v1 "github.com/ivannovak/glide/pkg/plugin/sdk/v1"
)

// GRPCPlugin implements the gRPC GlidePluginServer interface
type GRPCPlugin struct {
	*v1.BasePlugin
	detector *GoDetector
}

// NewGRPCPlugin creates a new gRPC-based Go plugin
func NewGRPCPlugin() *GRPCPlugin {
	metadata := &v1.PluginMetadata{
		Name:        "go",
		Version:     version.Version,
		Author:      "Glide Team",
		Description: "Go framework detector and command provider for Glide",
		Homepage:    "https://github.com/ivannovak/glide-plugin-go",
		License:     "MIT",
		Tags:        []string{"language", "go", "golang"},
		Aliases:     []string{"golang"},
		Namespaced:  false, // Commands don't need namespace (go build, not go:build)
	}

	p := &GRPCPlugin{
		BasePlugin: v1.NewBasePlugin(metadata),
		detector:   NewGoDetector(),
	}

	// Register all Go commands
	p.registerCommands()

	return p
}

// registerCommands registers all Go-related commands
func (p *GRPCPlugin) registerCommands() {
	commands := map[string]struct {
		cmd         string
		description string
		category    string
	}{
		"build": {
			cmd:         "go build ./...",
			description: "Build Go project",
			category:    "build",
		},
		"test": {
			cmd:         "go test ./...",
			description: "Run Go tests",
			category:    "test",
		},
		"test:v": {
			cmd:         "go test -v ./...",
			description: "Run tests with verbose output",
			category:    "test",
		},
		"test:race": {
			cmd:         "go test -race ./...",
			description: "Run tests with race detector",
			category:    "test",
		},
		"test:cover": {
			cmd:         "go test -cover ./...",
			description: "Run tests with coverage",
			category:    "test",
		},
		"run": {
			cmd:         "go run .",
			description: "Run Go application",
			category:    "run",
		},
		"fmt": {
			cmd:         "go fmt ./...",
			description: "Format Go code",
			category:    "format",
		},
		"vet": {
			cmd:         "go vet ./...",
			description: "Examine Go source code",
			category:    "lint",
		},
		"mod:tidy": {
			cmd:         "go mod tidy",
			description: "Add missing and remove unused modules",
			category:    "dependencies",
		},
		"mod:download": {
			cmd:         "go mod download",
			description: "Download modules to local cache",
			category:    "dependencies",
		},
		"mod:vendor": {
			cmd:         "go mod vendor",
			description: "Make vendored copy of dependencies",
			category:    "dependencies",
		},
		"generate": {
			cmd:         "go generate ./...",
			description: "Generate Go files",
			category:    "build",
		},
	}

	for name, def := range commands {
		cmdDef := def // Capture for closure
		handler := v1.NewSimpleCommand(
			&v1.CommandInfo{
				Name:        name,
				Description: cmdDef.description,
				Category:    cmdDef.category,
				Visibility:  "project-only", // Only show in Go projects
			},
			func(ctx context.Context, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
				return p.executeShellCommand(ctx, cmdDef.cmd, req)
			},
		)
		p.RegisterCommand(name, handler)
	}
}

// executeShellCommand runs a shell command and returns the response
func (p *GRPCPlugin) executeShellCommand(ctx context.Context, cmdStr string, req *v1.ExecuteRequest) (*v1.ExecuteResponse, error) {
	// Parse command string
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return &v1.ExecuteResponse{
			Success:  false,
			ExitCode: 1,
			Error:    "empty command",
		}, nil
	}

	// Append any additional args from the request
	parts = append(parts, req.Args...)

	// Create command
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = req.WorkDir
	if cmd.Dir == "" {
		cmd.Dir = "."
	}

	// Set environment - start with parent environment
	cmd.Env = os.Environ()
	// Override/add custom environment variables
	for k, v := range req.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Execute
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &v1.ExecuteResponse{
				Success:  false,
				ExitCode: 1,
				Error:    err.Error(),
			}, nil
		}
	}

	return &v1.ExecuteResponse{
		Success:  exitCode == 0,
		ExitCode: int32(exitCode),
		Stdout:   output,
		Stderr:   []byte{},
	}, nil
}

// DetectContext implements context detection for Go projects
func (p *GRPCPlugin) DetectContext(ctx context.Context, req *v1.ContextRequest) (*v1.ContextResponse, error) {
	// Use the detector to check if this is a Go project
	projectRoot := req.ProjectRoot
	if projectRoot == "" {
		projectRoot = req.WorkingDir
	}

	// Run detection
	data, err := p.detector.Detect(ctx, projectRoot)
	if err != nil {
		return &v1.ContextResponse{
			ExtensionName: "go",
			Detected:      false,
		}, nil
	}

	// If not detected, return early
	if data == nil {
		return &v1.ContextResponse{
			ExtensionName: "go",
			Detected:      false,
		}, nil
	}

	// Convert data map to response
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return &v1.ContextResponse{
			ExtensionName: "go",
			Detected:      false,
		}, nil
	}

	detected, _ := dataMap["detected"].(bool)
	if !detected {
		return &v1.ContextResponse{
			ExtensionName: "go",
			Detected:      false,
		}, nil
	}

	// Build response
	resp := &v1.ContextResponse{
		ExtensionName: "go",
		Detected:      true,
		Metadata:      make(map[string]string),
		Frameworks:    []string{},
		Tools:         []string{},
	}

	// Convert metadata
	for k, v := range dataMap {
		if k == "detected" || k == "commands" {
			continue
		}
		if str, ok := v.(string); ok {
			resp.Metadata[k] = str
		} else {
			resp.Metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	// Extract version
	if goVersion, ok := dataMap["go_version"].(string); ok {
		resp.Version = goVersion
	} else if version, ok := dataMap["version"].(string); ok {
		resp.Version = version
	}

	// Add Go framework indicator
	resp.Frameworks = append(resp.Frameworks, "go")

	// Check for tools
	if hasTools, ok := dataMap["has_dev_tools"].(bool); ok && hasTools {
		resp.Tools = append(resp.Tools, "golangci-lint", "goreleaser")
	}

	return resp, nil
}

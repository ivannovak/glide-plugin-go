package plugin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivannovak/glide/v3/pkg/plugin/sdk"
)

// GoDetector detects Go projects and implements sdk.ContextExtension
type GoDetector struct {
	base            *sdk.BaseFrameworkDetector
	enableWorkspace bool
	enableTools     bool
}

// NewGoDetector creates a new Go detector
func NewGoDetector() *GoDetector {
	base := sdk.NewBaseFrameworkDetector(sdk.FrameworkInfo{
		Name: "go",
		Type: "language",
	})

	// Set detection patterns
	base.SetPatterns(sdk.DetectionPatterns{
		RequiredFiles: []string{"go.mod"},
		OptionalFiles: []string{"go.sum", "go.work"},
		Directories:   []string{"vendor"},
		Extensions:    []string{".go"},
	})

	// Set default commands
	base.SetCommands(map[string]sdk.CommandDefinition{
		"build": {
			Cmd:         "go build ./...",
			Description: "Build Go project",
			Category:    "build",
		},
		"test": {
			Cmd:         "go test ./...",
			Description: "Run Go tests",
			Category:    "test",
		},
		"test:v": {
			Cmd:         "go test -v ./...",
			Description: "Run Go tests with verbose output",
			Category:    "test",
		},
		"test:race": {
			Cmd:         "go test -race ./...",
			Description: "Run Go tests with race detector",
			Category:    "test",
		},
		"test:cover": {
			Cmd:         "go test -cover ./...",
			Description: "Run Go tests with coverage",
			Category:    "test",
		},
		"run": {
			Cmd:         "go run .",
			Description: "Run Go application",
			Category:    "run",
		},
		"fmt": {
			Cmd:         "go fmt ./...",
			Description: "Format Go code",
			Category:    "format",
		},
		"vet": {
			Cmd:         "go vet ./...",
			Description: "Examine Go source code",
			Category:    "lint",
		},
		"mod:tidy": {
			Cmd:         "go mod tidy",
			Description: "Add missing and remove unused modules",
			Category:    "dependencies",
		},
		"mod:download": {
			Cmd:         "go mod download",
			Description: "Download modules to local cache",
			Category:    "dependencies",
		},
		"mod:vendor": {
			Cmd:         "go mod vendor",
			Description: "Make vendored copy of dependencies",
			Category:    "dependencies",
		},
		"generate": {
			Cmd:         "go generate ./...",
			Description: "Generate Go files",
			Category:    "build",
		},
	})

	return &GoDetector{
		base:            base,
		enableWorkspace: true,
		enableTools:     true,
	}
}

// SetEnableWorkspace sets whether workspace detection is enabled
func (d *GoDetector) SetEnableWorkspace(enabled bool) {
	d.enableWorkspace = enabled
}

// SetEnableTools sets whether tool detection is enabled
func (d *GoDetector) SetEnableTools(enabled bool) {
	d.enableTools = enabled
}

// Name returns the unique identifier for this extension
func (d *GoDetector) Name() string {
	return "go"
}

// Detect performs Go-specific detection and implements sdk.ContextExtension
func (d *GoDetector) Detect(ctx context.Context, projectPath string) (interface{}, error) {
	// First use base detection
	result, err := d.base.Detect(projectPath)
	if err != nil {
		return nil, err
	}
	if !result.Detected {
		return nil, nil // Not a Go project
	}

	// Build extension data
	data := map[string]interface{}{
		"detected":   true,
		"framework":  "go",
		"type":       "language",
		"confidence": result.Confidence,
		"commands":   result.Commands,
	}

	// Enhance with Go-specific detection
	goModPath := filepath.Join(projectPath, "go.mod")
	if version, err := d.detectGoVersion(goModPath); err == nil {
		data["version"] = version
		data["go_version"] = version
		data["module"] = d.detectModuleName(goModPath)
	}

	// Check for workspace (if enabled)
	if d.enableWorkspace {
		if _, err := os.Stat(filepath.Join(projectPath, "go.work")); err == nil {
			data["workspace"] = true
		}
	}

	// Check for common Go tools (if enabled)
	if d.enableTools && d.hasGoTools(projectPath) {
		data["has_dev_tools"] = true
		// Adjust confidence
		if conf, ok := data["confidence"].(int); ok {
			data["confidence"] = min(100, conf+10)
		}
	}

	return data, nil
}

// Merge combines this extension's data with existing extension data
func (d *GoDetector) Merge(existing interface{}, new interface{}) (interface{}, error) {
	// For Go detector, prefer new data over existing
	if new != nil {
		return new, nil
	}
	return existing, nil
}

// detectGoVersion extracts Go version from go.mod
func (d *GoDetector) detectGoVersion(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go "), nil
		}
	}

	return "", fmt.Errorf("go version not found in go.mod")
}

// detectModuleName extracts module name from go.mod
func (d *GoDetector) detectModuleName(goModPath string) string {
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}

	return ""
}

// hasGoTools checks for common Go development tools
func (d *GoDetector) hasGoTools(projectPath string) bool {
	toolFiles := []string{
		".golangci.yml",
		".golangci.yaml",
		".goreleaser.yml",
		".goreleaser.yaml",
		"Makefile", // Often used for Go projects
	}

	for _, file := range toolFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			return true
		}
	}

	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package plugin

import (
	"github.com/ivannovak/glide-plugin-go/pkg/version"
	"github.com/ivannovak/glide/v2/pkg/plugin"
	"github.com/ivannovak/glide/v2/pkg/plugin/sdk"
	"github.com/spf13/cobra"
)

// GoPlugin implements the SDK Plugin interfaces for Go framework detection
type GoPlugin struct {
	detector *GoDetector
}

// New creates a new Go plugin instance
func New() *GoPlugin {
	return &GoPlugin{
		detector: NewGoDetector(),
	}
}

// NewGoPlugin creates a new Go plugin instance (legacy name)
func NewGoPlugin() *GoPlugin {
	return New()
}

// Name returns the plugin identifier
func (p *GoPlugin) Name() string {
	return "go"
}

// Version returns the plugin version
func (p *GoPlugin) Version() string {
	return version.Version
}

// Description returns the plugin description
func (p *GoPlugin) Description() string {
	return "Go framework detector for Glide"
}

// Register adds plugin commands to the command tree
func (p *GoPlugin) Register(root *cobra.Command) error {
	// Framework detector plugins don't register commands
	// They only provide framework detection capabilities
	return nil
}

// Configure allows plugin-specific configuration
func (p *GoPlugin) Configure(config map[string]interface{}) error {
	// Go plugin doesn't require specific configuration
	return nil
}

// Metadata returns plugin information
func (p *GoPlugin) Metadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "go",
		Version:     version.Version,
		Author:      "Glide Team",
		Description: "Go framework detector for Glide",
		Aliases:     []string{"golang"},
		Commands:    []plugin.CommandInfo{},
		BuildTags:   []string{},
		ConfigKeys:  []string{"go"},
	}
}

// ProvideContext returns the context extension for Go detection
func (p *GoPlugin) ProvideContext() sdk.ContextExtension {
	return p.detector
}

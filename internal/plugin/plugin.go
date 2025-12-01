package plugin

import (
	"context"

	"github.com/ivannovak/glide-plugin-go/pkg/version"
	"github.com/ivannovak/glide/v3/pkg/plugin/sdk/v2"
)

// Config defines the plugin's type-safe configuration.
// Users configure this in .glide.yml under plugins.go
type Config struct {
	// EnableWorkspace enables Go workspace detection
	EnableWorkspace bool `json:"enableWorkspace" yaml:"enableWorkspace"`

	// EnableTools enables detection of Go development tools
	EnableTools bool `json:"enableTools" yaml:"enableTools"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		EnableWorkspace: true,
		EnableTools:     true,
	}
}

// GoPlugin implements the SDK v2 Plugin interface for Go framework detection
type GoPlugin struct {
	v2.BasePlugin[Config]
	detector *GoDetector
}

// New creates a new Go plugin instance
func New() *GoPlugin {
	p := &GoPlugin{
		detector: NewGoDetector(),
	}
	return p
}

// Metadata returns plugin information
func (p *GoPlugin) Metadata() v2.Metadata {
	return v2.Metadata{
		Name:        "go",
		Version:     version.Version,
		Author:      "Glide Team",
		Description: "Go framework detector for Glide",
		License:     "MIT",
		Homepage:    "https://github.com/ivannovak/glide-plugin-go",
		Tags:        []string{"language", "go", "golang", "detector"},
	}
}

// Configure is called with the type-safe configuration
func (p *GoPlugin) Configure(ctx context.Context, config Config) error {
	if err := p.BasePlugin.Configure(ctx, config); err != nil {
		return err
	}

	// Apply configuration to detector
	cfg := p.Config()
	p.detector.SetEnableWorkspace(cfg.EnableWorkspace)
	p.detector.SetEnableTools(cfg.EnableTools)

	return nil
}

// Commands returns the list of commands this plugin provides.
// Note: This is a framework detector plugin, so it doesn't provide CLI commands.
// Commands are dynamically provided based on detected context.
func (p *GoPlugin) Commands() []v2.Command {
	return []v2.Command{}
}

// Init is called once after plugin load
func (p *GoPlugin) Init(ctx context.Context) error {
	return nil
}

// HealthCheck returns nil if the plugin is healthy
func (p *GoPlugin) HealthCheck(ctx context.Context) error {
	return nil
}

package plugin

import (
	"github.com/ivannovak/glide-plugin-go/pkg/version"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
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
		Description: "Go framework detector for Glide",
		Homepage:    "https://github.com/ivannovak/glide-plugin-go",
		License:     "MIT",
		Tags:        []string{"language", "go", "golang", "detector"},
		Aliases:     []string{"golang"},
		Namespaced:  false,
	}

	p := &GRPCPlugin{
		BasePlugin: v1.NewBasePlugin(metadata),
		detector:   NewGoDetector(),
	}

	// Note: This plugin only provides framework detection, not commands
	// Commands are handled by Glide's core CLI based on detected context

	return p
}


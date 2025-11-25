package main

import (
	"log"
	"os"

	"github.com/ivannovak/glide-plugin-go/internal/plugin"
	v1 "github.com/ivannovak/glide/v2/pkg/plugin/sdk/v1"
)

func main() {
	// Create the plugin instance
	p := plugin.NewGRPCPlugin()

	// Start the plugin server
	if err := v1.RunPlugin(p); err != nil {
		log.Printf("Plugin server error: %v", err)
		os.Exit(1)
	}
}

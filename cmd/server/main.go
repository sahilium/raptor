// Command server is the Raptor MCP server entrypoint.
//
// Raptor exposes a set of structured infrastructure tools over the Model
// Context Protocol (MCP), allowing AI assistants to safely interact with
// remote backend infrastructure via SSH without unrestricted shell access.
//
// The server communicates over stdio using the mcp-go library and is intended
// to be launched as a subprocess by an MCP-compatible client such as Claude
// Desktop or a custom agent.
//
// Configuration is provided entirely through environment variables.
// See internal/config for the full list.
package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/tools/deployment"
	"github.com/sahilium/raptor/internal/tools/health"
	"github.com/sahilium/raptor/internal/tools/logs"
	"github.com/sahilium/raptor/internal/tools/release"
	"github.com/sahilium/raptor/internal/tools/service"
)

// version is the current Raptor server version.  It is embedded here rather
// than derived from build flags so that the value is visible in the MCP
// server info block returned during capability negotiation.
const version = "0.1.0"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "raptor: configuration error: %v\n", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"Raptor",
		version,
		server.WithToolCapabilities(true),
	)

	// Register tool groups.  Each package exposes a single Register function
	// that adds all tools belonging to that domain.
	deployment.Register(s, cfg)
	health.Register(s, cfg)
	logs.Register(s, cfg)
	release.Register(s, cfg)
	service.Register(s, cfg)

	// ServeStdio blocks until the client disconnects or an unrecoverable
	// error occurs.  All tool dispatch happens inside the mcp-go event loop.
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "raptor: server error: %v\n", err)
		os.Exit(1)
	}
}

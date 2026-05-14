// Package release provides Raptor MCP tools for querying release information
// on a remote host.
//
// Tools exposed:
//   - release_current — identify the currently deployed release version
//   - release_list    — enumerate all available release versions
package release

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/ssh"
)

// Register adds the release.current and release.list tools to the provided
// MCP server handle.
func Register(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	registerCurrent(s, cfg)
	registerList(s, cfg)
}

// registerCurrent wires up the release_current tool.
func registerCurrent(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"release_current",
		mcp.WithDescription(
			"Return the version string of the currently active release on the remote host. "+
				"The version is read from the file specified by RAPTOR_CURRENT_RELEASE_FILE.",
		),
	)

	s.AddTool(tool, makeCurrentHandler(cfg))
}

// registerList wires up the release_list tool.
func registerList(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"release_list",
		mcp.WithDescription(
			"List all release versions present in the releases directory on the remote host. "+
				"Versions are returned newest-first based on directory modification time.",
		),
	)

	s.AddTool(tool, makeListHandler(cfg))
}

// makeCurrentHandler returns the ToolHandlerFunc for release_current.
func makeCurrentHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf("cat %s", cfg.CurrentReleaseFile)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("could not read current release from %s: %v", cfg.CurrentReleaseFile, err),
			), nil
		}

		version := strings.TrimSpace(out)
		if version == "" {
			return mcp.NewToolResultText("No current release version found."), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Current release: %s", version)), nil
	}
}

// makeListHandler returns the ToolHandlerFunc for release_list.
func makeListHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		// Sort by modification time (newest first) and show only directory names.
		cmd := fmt.Sprintf(
			"ls -1t %s 2>/dev/null || echo ''",
			cfg.ReleasesDir,
		)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("could not list releases: %v\nOutput: %s", err, out)), nil
		}

		out = strings.TrimSpace(out)
		if out == "" {
			return mcp.NewToolResultText("No releases found."), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Available releases (newest first):\n%s", out)), nil
	}
}

// Package health provides the Raptor MCP health-check tool.
//
// The health.check tool runs a lightweight probe on all Docker Compose services
// defined on the remote host and reports their status in a human-readable form
// suitable for consumption by an AI assistant.
package health

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/ssh"
)

// Register adds the health.check tool to the provided MCP server handle.
// The tool opens a fresh SSH connection on every invocation so that transient
// network faults are reported directly to the caller rather than silently
// retried.
func Register(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"health_check",
		mcp.WithDescription(
			"Check the health of all running Docker Compose services on the remote host. "+
				"Returns the live status of each container.",
		),
	)

	s.AddTool(tool, makeHandler(cfg))
}

// makeHandler returns the ToolHandlerFunc for health.check, closing over cfg.
func makeHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		// docker compose ps provides a concise, human-readable status table.
		cmd := fmt.Sprintf(
			"docker compose -f %s ps --format 'table {{.Name}}\t{{.Status}}\t{{.Ports}}'",
			cfg.DockerComposeFile,
		)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("health check failed: %v\nOutput: %s", err, out)), nil
		}

		if out == "" {
			return mcp.NewToolResultText("No running services found."), nil
		}
		return mcp.NewToolResultText(out), nil
	}
}

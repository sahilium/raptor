// Package service provides Raptor MCP tools for querying and restarting Docker
// Compose services on a remote host.
//
// Tools exposed:
//   - service_status  — report the live status of one or all services
//   - service_restart — restart a named service
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/ssh"
)

// Register adds the service.status and service.restart tools to the provided
// MCP server handle.
func Register(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	registerStatus(s, cfg)
	registerRestart(s, cfg)
}

// registerStatus wires up the service_status tool.
func registerStatus(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"service_status",
		mcp.WithDescription(
			"Report the live status of one or all Docker Compose services on the remote host. "+
				"When 'service' is omitted all services are reported.",
		),
		mcp.WithString(
			"service",
			mcp.Description("Name of a specific Docker Compose service to inspect. Omit to list all services."),
		),
	)

	s.AddTool(tool, makeStatusHandler(cfg))
}

// registerRestart wires up the service_restart tool.
func registerRestart(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"service_restart",
		mcp.WithDescription(
			"Restart a named Docker Compose service on the remote host. "+
				"The restart is performed with 'docker compose restart' which preserves the "+
				"existing container configuration.",
		),
		mcp.WithString(
			"service",
			mcp.Required(),
			mcp.Description("Name of the Docker Compose service to restart."),
		),
	)

	s.AddTool(tool, makeRestartHandler(cfg))
}

// makeStatusHandler returns the ToolHandlerFunc for service_status.
func makeStatusHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		service := strings.TrimSpace(mcp.ParseString(req, "service", ""))

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		var cmd string
		if service != "" {
			cmd = fmt.Sprintf(
				"docker compose -f %s ps --format 'table {{.Name}}\t{{.Status}}\t{{.Ports}}' %s",
				cfg.DockerComposeFile,
				service,
			)
		} else {
			cmd = fmt.Sprintf(
				"docker compose -f %s ps --format 'table {{.Name}}\t{{.Status}}\t{{.Ports}}'",
				cfg.DockerComposeFile,
			)
		}

		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("status check failed: %v\nOutput: %s", err, out)), nil
		}

		if out == "" {
			return mcp.NewToolResultText("No running services found."), nil
		}
		return mcp.NewToolResultText(out), nil
	}
}

// makeRestartHandler returns the ToolHandlerFunc for service_restart.
func makeRestartHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		service, err := req.RequireString("service")
		if err != nil {
			return mcp.NewToolResultError("parameter 'service' is required"), nil
		}
		service = strings.TrimSpace(service)
		if service == "" {
			return mcp.NewToolResultError("parameter 'service' must not be empty"), nil
		}

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf(
			"docker compose -f %s restart %s",
			cfg.DockerComposeFile,
			service,
		)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("restart failed: %v\nOutput: %s", err, out)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Service %q restarted successfully.\n%s", service, out)), nil
	}
}

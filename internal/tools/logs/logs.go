// Package logs provides the Raptor MCP log-tailing tool.
//
// The logs.tail tool fetches the most recent N log lines from a named Docker
// Compose service running on the remote host and returns them as plain text.
package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/ssh"
)

const (
	// defaultLines is the number of log lines returned when the caller does
	// not specify an explicit value.
	defaultLines = 50

	// maxLines is the upper bound accepted from callers.  Values above this
	// are clamped to prevent excessively large responses.
	maxLines = 1000
)

// Register adds the logs.tail tool to the provided MCP server handle.
func Register(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"logs_tail",
		mcp.WithDescription(
			"Fetch the most recent log lines from a Docker Compose service running on the remote host. "+
				"Returns plain-text log output suitable for debugging.",
		),
		mcp.WithString(
			"service",
			mcp.Required(),
			mcp.Description("Name of the Docker Compose service whose logs to fetch."),
		),
		mcp.WithNumber(
			"lines",
			mcp.Description(fmt.Sprintf(
				"Number of log lines to return. Defaults to %d, capped at %d.",
				defaultLines, maxLines,
			)),
		),
	)

	s.AddTool(tool, makeHandler(cfg))
}

// makeHandler returns the ToolHandlerFunc for logs.tail, closing over cfg.
func makeHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		service, err := req.RequireString("service")
		if err != nil {
			return mcp.NewToolResultError("parameter 'service' is required"), nil
		}
		service = strings.TrimSpace(service)
		if service == "" {
			return mcp.NewToolResultError("parameter 'service' must not be empty"), nil
		}

		lines := int(mcp.ParseFloat64(req, "lines", float64(defaultLines)))
		if lines <= 0 {
			lines = defaultLines
		}
		if lines > maxLines {
			lines = maxLines
		}

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf(
			"docker compose -f %s logs --no-color --tail=%d %s",
			cfg.DockerComposeFile,
			lines,
			service,
		)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("log fetch failed: %v\nOutput: %s", err, out)), nil
		}

		if out == "" {
			return mcp.NewToolResultText(fmt.Sprintf("No log output found for service %q.", service)), nil
		}
		return mcp.NewToolResultText(out), nil
	}
}

// Package deployment provides Raptor MCP tools for managing deployments on a
// remote host.
//
// Tools exposed:
//   - deployment_deploy   — trigger a versioned deployment via a remote script
//   - deployment_rollback — roll back to a previous release via a remote script
//   - deployment_history  — display the deployment history log
//
// Mutating operations (deploy, rollback) are intentionally kept separate from
// read-only tools so that callers can distinguish actions with side-effects.
package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sahilium/raptor/internal/config"
	"github.com/sahilium/raptor/internal/ssh"
)

// Register adds all deployment tools to the provided MCP server handle.
func Register(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	registerDeploy(s, cfg)
	registerRollback(s, cfg)
	registerHistory(s, cfg)
}

// registerDeploy wires up the deployment_deploy tool.
func registerDeploy(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"deployment_deploy",
		mcp.WithDescription(
			"Deploy a specific release version to the remote host by invoking the configured "+
				"deploy script with the target version as an argument. "+
				"The script is expected to pull the image, perform a rolling restart, "+
				"and record the deployment in the history log.",
		),
		mcp.WithString(
			"version",
			mcp.Required(),
			mcp.Description(
				"The release version to deploy (e.g. 'v1.4.2' or a Docker image tag). "+
					"This value is passed verbatim to the remote deploy script.",
			),
		),
	)

	s.AddTool(tool, makeDeployHandler(cfg))
}

// registerRollback wires up the deployment_rollback tool.
func registerRollback(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"deployment_rollback",
		mcp.WithDescription(
			"Roll back the remote host to a previously deployed release version by invoking "+
				"the configured rollback script with the target version as an argument.",
		),
		mcp.WithString(
			"version",
			mcp.Required(),
			mcp.Description(
				"The release version to roll back to. "+
					"Use release_list to discover available versions.",
			),
		),
	)

	s.AddTool(tool, makeRollbackHandler(cfg))
}

// registerHistory wires up the deployment_history tool.
func registerHistory(s interface {
	AddTool(mcp.Tool, server.ToolHandlerFunc)
}, cfg *config.ServerConfig) {
	tool := mcp.NewTool(
		"deployment_history",
		mcp.WithDescription(
			"Return the deployment history log from the remote host. "+
				"Each line represents one deployment event.",
		),
		mcp.WithNumber(
			"lines",
			mcp.Description("Maximum number of history entries to return. Defaults to 20."),
		),
	)

	s.AddTool(tool, makeHistoryHandler(cfg))
}

// makeDeployHandler returns the ToolHandlerFunc for deployment_deploy.
func makeDeployHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		version, err := req.RequireString("version")
		if err != nil {
			return mcp.NewToolResultError("parameter 'version' is required"), nil
		}
		version = strings.TrimSpace(version)
		if version == "" {
			return mcp.NewToolResultError("parameter 'version' must not be empty"), nil
		}
		if err := validateVersionString(version); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf("bash %s %s", cfg.DeployScript, version)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("deployment of version %q failed: %v\nOutput:\n%s", version, err, out),
			), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Deployment of version %q completed.\n\n%s", version, out),
		), nil
	}
}

// makeRollbackHandler returns the ToolHandlerFunc for deployment_rollback.
func makeRollbackHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		version, err := req.RequireString("version")
		if err != nil {
			return mcp.NewToolResultError("parameter 'version' is required"), nil
		}
		version = strings.TrimSpace(version)
		if version == "" {
			return mcp.NewToolResultError("parameter 'version' must not be empty"), nil
		}
		if err := validateVersionString(version); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf("bash %s %s", cfg.RollbackScript, version)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("rollback to version %q failed: %v\nOutput:\n%s", version, err, out),
			), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Rollback to version %q completed.\n\n%s", version, out),
		), nil
	}
}

// makeHistoryHandler returns the ToolHandlerFunc for deployment_history.
func makeHistoryHandler(cfg *config.ServerConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lines := int(mcp.ParseFloat64(req, "lines", 20))
		if lines <= 0 {
			lines = 20
		}
		if lines > 500 {
			lines = 500
		}

		client, err := ssh.New(&cfg.SSH)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to connect to host: %v", err)), nil
		}
		defer client.Close()

		cmd := fmt.Sprintf(
			"tail -n %d %s 2>/dev/null || echo 'No deployment history found.'",
			lines,
			cfg.DeployHistoryFile,
		)
		out, err := client.Execute(ctx, cmd)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("could not read deployment history: %v\nOutput: %s", err, out),
			), nil
		}

		if out == "" {
			return mcp.NewToolResultText("Deployment history is empty."), nil
		}
		return mcp.NewToolResultText(out), nil
	}
}

// validateVersionString performs a conservative sanity check on a version
// string before it is interpolated into a remote shell command.  Only
// alphanumeric characters, dots, hyphens, underscores, and forward slashes
// are permitted, matching common semver and Docker image tag patterns.
//
// This is a defence-in-depth measure; the primary security boundary is that
// Raptor never exposes arbitrary shell execution.
func validateVersionString(v string) error {
	for _, r := range v {
		if !isAllowed(r) {
			return fmt.Errorf(
				"version string %q contains invalid character %q; "+
					"only alphanumeric characters, dots, hyphens, underscores, and slashes are allowed",
				v, r,
			)
		}
	}
	return nil
}

// isAllowed reports whether rune r is safe for inclusion in a version string.
func isAllowed(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '.' || r == '-' || r == '_' || r == '/'
}

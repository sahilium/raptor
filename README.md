# Raptor

![Go Version](https://img.shields.io/github/go-mod/go-version/sahilium/raptor?style=for-the-badge)
![Release](https://img.shields.io/github/v/release/sahilium/raptor?style=for-the-badge)
![License](https://img.shields.io/github/license/sahilium/raptor?style=for-the-badge)
![Build](https://img.shields.io/github/actions/workflow/status/sahilium/raptor/release.yml?style=for-the-badge)

Raptor is a lightweight infrastructure control plane built on MCP (Model Context Protocol).

It allows AI assistants to safely interact with backend infrastructure using structured tools
instead of unrestricted shell access.

Raptor is designed for small teams running simple production setups:

- single EC2 instances or VPS hosts
- Docker Compose deployments
- GitHub Actions pipelines
- staging and production environments
- versioned releases and rollbacks

Instead of exposing a raw terminal, Raptor exposes high-level infrastructure operations:

- deploy a release
- rollback a deployment
- fetch logs
- check service health
- inspect running versions
- view deployment history

---

## Quick Start

### 1. Install

#### Option A: Download Binary
Grab the latest stable binary for your platform from the [GitHub Releases](https://github.com/sahilium/raptor/releases) page.

#### Option B: Go Install
```bash
go install github.com/sahilium/raptor/cmd/raptor@latest
```

#### Option C: Build from Source
```bash
git clone https://github.com/sahilium/raptor.git
cd raptor
go build -o raptor ./cmd/raptor
```

### 2. Configure
Set the required SSH environment variables (see `.env.example` for all options):
```bash
export RAPTOR_SSH_HOST="prod.example.com"
export RAPTOR_SSH_USER="deploy"
```

### 3. Run
Raptor communicates over `stdio`. You can run it directly to test, or add it to an MCP client like **Claude Desktop**:
```bash
./raptor
```

---

## Why does this exist?

Most infrastructure tooling is either:

- too low-level and dangerous
- or built for massive Kubernetes-heavy organizations

Raptor focuses on a simpler use case:
small backend services that still need safe deployments, visibility, and operational tooling.

The goal is to make infrastructure:

- observable
- explainable
- automatable
- AI-friendly

without giving an LLM unrestricted SSH access to production servers.

---

## Tools

All tools communicate over stdio using the Model Context Protocol.
Each tool name maps directly to the MCP `tools/call` method.

### Deployment

| Tool                  | Description                                                   |
|-----------------------|---------------------------------------------------------------|
| `deployment_deploy`   | Deploy a specific release version via a remote deploy script. |
| `deployment_rollback` | Roll back to a previous release via a remote rollback script. |
| `deployment_history`  | Return the last N lines of the deployment history log.        |

### Service

| Tool               | Description                                                |
|--------------------|------------------------------------------------------------|
| `service_status`   | Report live status of one or all Docker Compose services.  |
| `service_restart`  | Restart a named Docker Compose service.                    |

### Logs

| Tool        | Description                                                          |
|-------------|----------------------------------------------------------------------|
| `logs_tail` | Fetch the most recent N log lines from a named Compose service.      |

### Health

| Tool           | Description                                                        |
|----------------|--------------------------------------------------------------------|
| `health_check` | Run a liveness probe across all Docker Compose services.           |

### Release

| Tool              | Description                                             |
|-------------------|---------------------------------------------------------|
| `release_current` | Return the currently active release version string.     |
| `release_list`    | List all available release versions, newest first.      |

---

## Design Principles

### Structured operations over shell access

Raptor models infrastructure concepts directly:

- deployments
- releases
- services
- environments

instead of exposing arbitrary command execution.

---

### Safe by default

Raptor is designed around constrained operations:

- validated inputs (version strings are character-whitelisted before interpolation)
- explicit, named actions
- rollback-aware workflows
- no raw shell passthrough

The AI should never be able to execute unrestricted shell commands.

---

### Small-team focused

Raptor is intentionally optimized for:

- Docker Compose
- EC2 / VPS deployments
- simple backend stacks
- lightweight infrastructure

not large-scale orchestration systems.

---

## Tech Stack

- **Go** — server runtime
- **mcp-go** — MCP protocol implementation (stdio transport)
- **golang.org/x/crypto/ssh** — SSH client for remote execution
- **Docker Compose** — target container runtime on remote hosts

---

## Project Layout

```
raptor/
  cmd/
    server/
      main.go                     # Server entrypoint; registers all tool groups
  internal/
    config/
      config.go                   # Environment-variable configuration loader
    ssh/
      client.go                   # SSH client and remote executor
    tools/
      deployment/
        deployment.go             # deploy, rollback, history tools
      health/
        health.go                 # health_check tool
      logs/
        logs.go                   # logs_tail tool
      release/
        release.go                # release_current, release_list tools
      service/
        service.go                # service_status, service_restart tools
```

---

## Configuration

All configuration is provided via environment variables. No config file is required.

| Variable                      | Required | Default                              | Description                                      |
|-------------------------------|----------|--------------------------------------|--------------------------------------------------|
| `RAPTOR_SSH_HOST`             | Yes      | —                                    | Remote host address (hostname or IP).            |
| `RAPTOR_SSH_USER`             | Yes      | —                                    | SSH login username.                              |
| `RAPTOR_SSH_PORT`             | No       | `22`                                 | SSH port.                                        |
| `RAPTOR_SSH_IDENTITY_FILE`    | No       | —                                    | Path to SSH private key (preferred over password).|
| `RAPTOR_SSH_PASSWORD`         | No       | —                                    | SSH password (used when no key file is set).     |
| `RAPTOR_SSH_KNOWN_HOSTS`      | No       | —                                    | Path to known_hosts for host-key verification.   |
| `RAPTOR_COMPOSE_FILE`         | No       | `/opt/app/docker-compose.yml`        | Path to docker-compose file on the remote host.  |
| `RAPTOR_DEPLOY_SCRIPT`        | No       | `/opt/app/scripts/deploy.sh`         | Absolute path to deploy script on remote host.   |
| `RAPTOR_ROLLBACK_SCRIPT`      | No       | `/opt/app/scripts/rollback.sh`       | Absolute path to rollback script on remote host. |
| `RAPTOR_RELEASES_DIR`         | No       | `/opt/app/releases`                  | Directory containing release sub-directories.    |
| `RAPTOR_CURRENT_RELEASE_FILE` | No       | `/opt/app/current`                   | File whose contents name the active release.     |
| `RAPTOR_DEPLOY_HISTORY_FILE`  | No       | `/opt/app/deploy-history.log`        | Newline-delimited deployment history log.        |

See `.env.example` for a template.

---

## Running

### Build

```sh
go build -o raptor ./cmd/server
```

### Run (stdio transport)

```sh
RAPTOR_SSH_HOST=prod.example.com \
RAPTOR_SSH_USER=deploy \
RAPTOR_SSH_IDENTITY_FILE=~/.ssh/id_ed25519 \
./raptor
```

The server communicates over stdin/stdout using the MCP protocol and is intended to be
launched as a subprocess by an MCP-compatible client (e.g. Claude Desktop, a custom agent).

### Claude Desktop example (`claude_desktop_config.json`)

```json
{
  "mcpServers": {
    "raptor": {
      "command": "/path/to/raptor",
      "env": {
        "RAPTOR_SSH_HOST": "prod.example.com",
        "RAPTOR_SSH_USER": "deploy",
        "RAPTOR_SSH_IDENTITY_FILE": "/home/user/.ssh/id_ed25519"
      }
    }
  }
}
```

---

## Remote Host Conventions

Raptor makes the following assumptions about the remote host layout (all paths are configurable):

- `/opt/app/docker-compose.yml` — the active Compose file
- `/opt/app/scripts/deploy.sh <version>` — deploy script; receives version as `$1`
- `/opt/app/scripts/rollback.sh <version>` — rollback script; receives version as `$1`
- `/opt/app/releases/` — directory of release versions (one sub-directory per release)
- `/opt/app/current` — plain-text file containing the active release version string
- `/opt/app/deploy-history.log` — append-only deployment event log

The deploy and rollback scripts are operator-provided and run under the SSH user's permissions.
Raptor does not dictate their implementation.

---

## Development & Testing

You can test Raptor locally by SSHing into your own machine using a mock environment.

### 1. Setup Mock Environment
This script creates a dummy app structure in `./test_env` and generates a `.env` file.
```bash
./scripts/setup-test.sh
source .env
```

### 2. Build
```bash
go build -o raptor ./cmd/server
```

### 3. Test with MCP Inspector
The inspector provides a web UI to manually trigger tools.
```bash
bunx @modelcontextprotocol/inspector ./raptor
```

### 4. Test with Claude CLI
To use Raptor directly in your terminal with Claude:
```bash
# Add to Claude
claude mcp add ./raptor

# Or run a one-off session
claude --mcp ./raptor
```

---

## Security

- Raptor never exposes arbitrary command execution. Every operation is a named, validated action.
- Version strings passed to deploy/rollback are character-whitelisted (alphanumeric, `.`, `-`, `_`, `/`).
- SSH host-key verification is enforced when `RAPTOR_SSH_KNOWN_HOSTS` is set.
- Without a known_hosts file, host-key checking is skipped — acceptable for controlled single-operator
  environments, but operators in shared or untrusted networks should always set `RAPTOR_SSH_KNOWN_HOSTS`.

---

## Status

Early development. Currently implemented:

- MCP server foundation (stdio transport)
- SSH execution layer
- Read-only infrastructure tools: `health_check`, `service_status`, `logs_tail`, `release_current`, `release_list`, `deployment_history`
- Mutating tools with input validation: `deployment_deploy`, `deployment_rollback`, `service_restart`

Planned:

- GitHub deployment integration
- Incident summaries
- Multi-environment support
- Structured audit log

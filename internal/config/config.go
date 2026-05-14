// Package config provides configuration primitives for the Raptor MCP server.
// All values are loaded from environment variables at startup; no config file is required.
package config

import (
	"errors"
	"os"
	"strconv"
)

// SSHConfig holds the parameters required to establish an SSH connection to a
// remote host.
type SSHConfig struct {
	// Host is the remote host address (hostname or IP).
	Host string

	// Port is the TCP port on which sshd is listening. Defaults to 22.
	Port int

	// User is the SSH login username.
	User string

	// IdentityFile is the path to the private-key file used for authentication.
	// Mutually exclusive with Password; IdentityFile takes precedence when both
	// are set.
	IdentityFile string

	// Password is the SSH password used for authentication when IdentityFile is
	// empty. Storing credentials in environment variables is acceptable for
	// controlled, single-operator setups.
	Password string

	// KnownHostsFile is the path to the known_hosts file used for host-key
	// verification. When empty, strict host-key checking is disabled.
	KnownHostsFile string
}

// ServerConfig holds top-level operational parameters for the Raptor server.
type ServerConfig struct {
	SSH SSHConfig

	// DockerComposeFile is the path on the remote host to the
	// docker-compose.yml (or docker-compose.yaml) used to manage services.
	DockerComposeFile string

	// DeployScript is the absolute path on the remote host of the shell script
	// that performs a deployment. The script receives the target version string
	// as its first positional argument.
	DeployScript string

	// RollbackScript is the absolute path on the remote host of the shell
	// script that performs a rollback. The script receives the target version
	// string as its first positional argument.
	RollbackScript string

	// ReleasesDir is the absolute path on the remote host of the directory
	// that contains per-release sub-directories. Each sub-directory name is
	// treated as a release version string.
	ReleasesDir string

	// CurrentReleaseFile is the absolute path on the remote host of the file
	// whose contents identify the currently active release version.
	CurrentReleaseFile string

	// DeployHistoryFile is the absolute path on the remote host of the file
	// that contains newline-delimited deployment history records.
	DeployHistoryFile string
}

// Load reads all Raptor configuration from environment variables and returns a
// fully-populated ServerConfig.  Required variables that are missing or empty
// cause an error to be returned.
//
// Environment variables:
//
//	RAPTOR_SSH_HOST              required  – remote host address
//	RAPTOR_SSH_PORT              optional  – defaults to 22
//	RAPTOR_SSH_USER              required  – SSH login username
//	RAPTOR_SSH_IDENTITY_FILE     optional  – path to private key
//	RAPTOR_SSH_PASSWORD          optional  – password (used when no key file)
//	RAPTOR_SSH_KNOWN_HOSTS       optional  – path to known_hosts
//	RAPTOR_COMPOSE_FILE          optional  – docker-compose file path on remote
//	RAPTOR_DEPLOY_SCRIPT         optional  – deploy script path on remote
//	RAPTOR_ROLLBACK_SCRIPT       optional  – rollback script path on remote
//	RAPTOR_RELEASES_DIR          optional  – releases directory on remote
//	RAPTOR_CURRENT_RELEASE_FILE  optional  – current-release marker file on remote
//	RAPTOR_DEPLOY_HISTORY_FILE   optional  – deployment history file on remote
func Load() (*ServerConfig, error) {
	host := os.Getenv("RAPTOR_SSH_HOST")
	if host == "" {
		return nil, errors.New("config: RAPTOR_SSH_HOST is required")
	}

	user := os.Getenv("RAPTOR_SSH_USER")
	if user == "" {
		return nil, errors.New("config: RAPTOR_SSH_USER is required")
	}

	port := 22
	if raw := os.Getenv("RAPTOR_SSH_PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil || p <= 0 || p > 65535 {
			return nil, errors.New("config: RAPTOR_SSH_PORT must be a valid port number")
		}
		port = p
	}

	cfg := &ServerConfig{
		SSH: SSHConfig{
			Host:           host,
			Port:           port,
			User:           user,
			IdentityFile:   os.Getenv("RAPTOR_SSH_IDENTITY_FILE"),
			Password:       os.Getenv("RAPTOR_SSH_PASSWORD"),
			KnownHostsFile: os.Getenv("RAPTOR_SSH_KNOWN_HOSTS"),
		},
		DockerComposeFile:  getEnvDefault("RAPTOR_COMPOSE_FILE", "/opt/app/docker-compose.yml"),
		DeployScript:       getEnvDefault("RAPTOR_DEPLOY_SCRIPT", "/opt/app/scripts/deploy.sh"),
		RollbackScript:     getEnvDefault("RAPTOR_ROLLBACK_SCRIPT", "/opt/app/scripts/rollback.sh"),
		ReleasesDir:        getEnvDefault("RAPTOR_RELEASES_DIR", "/opt/app/releases"),
		CurrentReleaseFile: getEnvDefault("RAPTOR_CURRENT_RELEASE_FILE", "/opt/app/current"),
		DeployHistoryFile:  getEnvDefault("RAPTOR_DEPLOY_HISTORY_FILE", "/opt/app/deploy-history.log"),
	}

	return cfg, nil
}

// getEnvDefault returns the value of the named environment variable, or
// fallback if the variable is unset or empty.
func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

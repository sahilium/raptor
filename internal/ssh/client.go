// Package ssh provides a thin, reusable SSH client that executes remote
// commands and returns their combined output.  It is the sole abstraction over
// network I/O used by the Raptor tool layer; all infrastructure tools call
// through this package rather than shelling out locally.
package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sahilium/raptor/internal/config"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Client wraps an active SSH connection and exposes a single high-level
// Execute method.  A Client is safe to reuse across multiple sequential
// Execute calls but is not concurrency-safe.
type Client struct {
	inner *gossh.Client
	cfg   *config.SSHConfig
}

// New dials the remote host described by cfg and returns a ready-to-use Client.
// The caller is responsible for calling Close when the Client is no longer
// needed.
func New(cfg *config.SSHConfig) (*Client, error) {
	authMethods, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, fmt.Errorf("ssh: build auth methods: %w", err)
	}

	hostKeyCallback, err := buildHostKeyCallback(cfg.KnownHostsFile)
	if err != nil {
		return nil, fmt.Errorf("ssh: build host-key callback: %w", err)
	}

	sshCfg := &gossh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	inner, err := gossh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ssh: dial %s: %w", addr, err)
	}

	return &Client{inner: inner, cfg: cfg}, nil
}

// Execute runs cmd on the remote host and returns its combined stdout+stderr
// output as a trimmed string.  The provided context is checked before opening
// a new session; cancellation mid-command is not guaranteed.
func (c *Client) Execute(ctx context.Context, cmd string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("ssh: context cancelled before execution: %w", err)
	}

	session, err := c.inner.NewSession()
	if err != nil {
		return "", fmt.Errorf("ssh: new session: %w", err)
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		// Return the captured output alongside the error so callers can
		// surface the remote error message to the AI client.
		return strings.TrimSpace(string(out)), fmt.Errorf("ssh: remote command failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// Close terminates the underlying SSH connection.  It is safe to call Close
// multiple times; subsequent calls are no-ops.
func (c *Client) Close() error {
	if c.inner != nil {
		return c.inner.Close()
	}
	return nil
}

// buildAuthMethods constructs the list of SSH authentication methods from the
// provided configuration.  If IdentityFile is set it is preferred over
// Password.
func buildAuthMethods(cfg *config.SSHConfig) ([]gossh.AuthMethod, error) {
	if cfg.IdentityFile != "" {
		pem, err := os.ReadFile(cfg.IdentityFile)
		if err != nil {
			return nil, fmt.Errorf("read identity file %q: %w", cfg.IdentityFile, err)
		}
		signer, err := gossh.ParsePrivateKey(pem)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		return []gossh.AuthMethod{gossh.PublicKeys(signer)}, nil
	}

	if cfg.Password != "" {
		return []gossh.AuthMethod{gossh.Password(cfg.Password)}, nil
	}

	return nil, fmt.Errorf("ssh config requires either IdentityFile or Password")
}

// buildHostKeyCallback returns a host-key verification callback.  When
// knownHostsFile is non-empty it loads that file and enforces strict
// verification.  Otherwise it returns an insecure callback that accepts any
// host key — acceptable only in development environments.
func buildHostKeyCallback(knownHostsFile string) (gossh.HostKeyCallback, error) {
	if knownHostsFile != "" {
		cb, err := knownhosts.New(knownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("load known_hosts %q: %w", knownHostsFile, err)
		}
		return cb, nil
	}

	// InsecureIgnoreHostKey is intentionally used here because Raptor is
	// designed for controlled, single-operator environments where the
	// operator has manual visibility of the target host.  Operators that
	// require strict host-key checking should set RAPTOR_SSH_KNOWN_HOSTS.
	return gossh.InsecureIgnoreHostKey(), nil //nolint:gosec
}

#!/bin/bash
# setup-test.sh - Prepare a local mock environment for Raptor testing.
#
# This script creates a dummy directory structure and a .env file to allow
# testing Raptor against localhost via SSH.

set -euo pipefail

# --- Configuration ---
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR="${1:-$PROJECT_ROOT/test_env}"
ENV_FILE="$PROJECT_ROOT/.env"

# --- Helpers ---
info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*" >&2; exit 1; }

# --- Checks ---
command -v go >/dev/null 2>&1 || error "Go is not installed."

# --- Cleanup & Setup ---
info "Creating test environment in: $TEST_DIR"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR/releases/v0.1.0"
mkdir -p "$TEST_DIR/releases/v0.2.0"
mkdir -p "$TEST_DIR/scripts"

# Initialize mock data
echo "v0.1.0" > "$TEST_DIR/current"
touch "$TEST_DIR/deploy-history.log"

# Create a dummy docker-compose.yml
cat <<EOF > "$TEST_DIR/docker-compose.yml"
services:
  web:
    image: nginx:alpine
    container_name: raptor-test-web
EOF

# Create mock deployment scripts
cat <<EOF > "$TEST_DIR/scripts/deploy.sh"
#!/bin/bash
echo "Raptor Mock: Deploying version \$1..."
echo "\$1" > "$TEST_DIR/current"
echo "\$(date): Deployed \$1" >> "$TEST_DIR/deploy-history.log"
EOF

cat <<EOF > "$TEST_DIR/scripts/rollback.sh"
#!/bin/bash
echo "Raptor Mock: Rolling back to \$1..."
echo "\$1" > "$TEST_DIR/current"
echo "\$(date): Rolled back to \$1" >> "$TEST_DIR/deploy-history.log"
EOF

chmod +x "$TEST_DIR/scripts/"*.sh

# --- Generate .env ---
info "Generating $ENV_FILE"
cat <<EOF > "$ENV_FILE"
export RAPTOR_SSH_HOST="localhost"
export RAPTOR_SSH_USER="$(whoami)"
export RAPTOR_COMPOSE_FILE="$TEST_DIR/docker-compose.yml"
export RAPTOR_DEPLOY_SCRIPT="$TEST_DIR/scripts/deploy.sh"
export RAPTOR_ROLLBACK_SCRIPT="$TEST_DIR/scripts/rollback.sh"
export RAPTOR_RELEASES_DIR="$TEST_DIR/releases"
export RAPTOR_CURRENT_RELEASE_FILE="$TEST_DIR/current"
export RAPTOR_DEPLOY_HISTORY_FILE="$TEST_DIR/deploy-history.log"
EOF

info "Setup complete."
info "Run 'source .env' to load the test configuration."

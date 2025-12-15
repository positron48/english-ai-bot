#!/bin/bash

# Universal AI Bot Deployment Script
set -e

# Configuration - can be overridden by environment variables
GITHUB_REPO="${GITHUB_REPO:-positron48/universal-ai-bot}"
APP_NAME="${APP_NAME:-universal-ai-bot}"
APP_DIR="${PWD}"  # Use current directory

# Load SERVICE_NAME from .env file
if [ -f "$APP_DIR/.env" ]; then
    SERVICE_NAME=$(grep '^SERVICE_NAME=' "$APP_DIR/.env" | cut -d'=' -f2)
fi
SERVICE_NAME="${SERVICE_NAME:-ai-bot}"

# Get latest release
get_latest_release() {
    curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | jq -r '.tag_name'
}

# Download and install binary
deploy() {
    echo "Deploying AI Bot..."
    
    # Create directories
    mkdir -p "$APP_DIR"/{bin,data,logs,configs}
    
    # Get version
    VERSION=$(get_latest_release)
    echo "Version: $VERSION"
    
    # Stop service before update
    if systemctl --user is-active $SERVICE_NAME >/dev/null 2>&1; then
        echo "Stopping service..."
        systemctl --user stop $SERVICE_NAME
    fi
    
    # Download binary
    BINARY_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$APP_NAME-linux_amd64"
    curl -L -o "$APP_DIR/bin/$APP_NAME" "$BINARY_URL"
    chmod +x "$APP_DIR/bin/$APP_NAME"
    
    # Create .env if not exists
    if [ ! -f "$APP_DIR/.env" ]; then
        curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/master/env.example" > "$APP_DIR/.env"
        echo "Created .env file. Edit it: nano $APP_DIR/.env"
    fi
    
    # Download configs
    echo "Downloading configs..."
    curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/master/configs/config.yaml" > "$APP_DIR/configs/config.yaml" 2>/dev/null || echo "Config not found"
    
    # Restart service (if already configured)
    if systemctl --user is-enabled $SERVICE_NAME >/dev/null 2>&1; then
        systemctl --user restart $SERVICE_NAME
    else
        echo "Service not configured. Run setup first."
        exit 1
    fi
    
    echo "Done! Status: systemctl --user status $SERVICE_NAME"
}

# Update
update() {
    echo "Updating AI Bot..."
    systemctl --user stop $SERVICE_NAME
    deploy
}

# Status
status() {
    systemctl --user status $SERVICE_NAME
}

# Logs
logs() {
    journalctl --user -u $SERVICE_NAME -f
}

case "${1:-deploy}" in
    deploy) deploy ;;
    update) update ;;
    status) status ;;
    logs) logs ;;
    *) echo "Usage: $0 [deploy|update|status|logs]" ;;
esac

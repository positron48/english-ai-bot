#!/bin/bash

# Universal AI Bot Deployment Script
set -e

# Configuration - can be overridden by environment variables
GITHUB_REPO="${GITHUB_REPO:-positron48/english-ai-bot}"
APP_NAME="${APP_NAME:-universal-ai-bot}"
APP_DIR="${PWD}"  # Use current directory

# Load variables from .env file
if [ -f "$APP_DIR/.env" ]; then
    # Source the .env file to load variables safely
    set -a  # automatically export all variables
    source "$APP_DIR/.env"
    set +a  # stop automatically exporting
fi
SERVICE_NAME="${SERVICE_NAME:-ai-bot}"

# Get latest release
get_latest_release() {
    local response
    if [ -n "$GITHUB_TOKEN" ]; then
        response=$(curl -s -w "\n%{http_code}" -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/$GITHUB_REPO/releases/latest")
    else
        response=$(curl -s -w "\n%{http_code}" "https://api.github.com/repos/$GITHUB_REPO/releases/latest")
    fi
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "200" ]; then
        echo "Error: GitHub API returned HTTP $http_code" >&2
        echo "$body" >&2
        echo "null"
        return 1
    fi
    
    echo "$body" | jq -r '.tag_name'
}

# Download and install binary
deploy() {
    echo "Deploying AI Bot..."
    
    # Create directories
    mkdir -p "$APP_DIR"/{bin,data,logs,configs}
    
    # Get version
    VERSION=$(get_latest_release)
    if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
        echo "Error: Could not get latest release version"
        echo "Check GITHUB_TOKEN if using private repository"
        exit 1
    fi
    echo "Version: $VERSION"
    
    # Stop service before update
    if systemctl --user is-active $SERVICE_NAME >/dev/null 2>&1; then
        echo "Stopping service..."
        systemctl --user stop $SERVICE_NAME
    fi
    
    # Download binary
    # For private repos, we need to use the API URL (not browser_download_url)
    if [ -n "$GITHUB_TOKEN" ]; then
        echo "Using GitHub API to download binary (private repository)"
        # Get the API URL for the asset (requires authentication)
        ASSET_URL=$(curl -s -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$VERSION" | jq -r ".assets[] | select(.name == \"$APP_NAME-linux_amd64\") | .url")
        
        if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
            echo "Error: Could not find asset $APP_NAME-linux_amd64 in release $VERSION"
            echo "Available assets:"
            curl -s -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$VERSION" | jq -r '.assets[].name'
            exit 1
        fi
        
        echo "Downloading asset via GitHub API..."
        HTTP_CODE=$(curl -L -s -o "$APP_DIR/bin/$SERVICE_NAME" -w "%{http_code}" -H "Authorization: token $GITHUB_TOKEN" -H "Accept: application/octet-stream" "$ASSET_URL")
    else
        BINARY_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$APP_NAME-linux_amd64"
        echo "Downloading from: $BINARY_URL"
        HTTP_CODE=$(curl -L -s -o "$APP_DIR/bin/$SERVICE_NAME" -w "%{http_code}" "$BINARY_URL")
    fi
    
    # Check HTTP status code
    if [ "$HTTP_CODE" != "200" ]; then
        echo "Error: Failed to download binary. HTTP code: $HTTP_CODE"
        echo "Response content:"
        cat "$APP_DIR/bin/$SERVICE_NAME"
        echo ""
        echo "Available assets in release $VERSION:"
        if [ -n "$GITHUB_TOKEN" ]; then
            curl -s -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$VERSION" | jq -r '.assets[].name' | head -10
        else
            curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$VERSION" | jq -r '.assets[].name' | head -10
        fi
        rm -f "$APP_DIR/bin/$SERVICE_NAME"
        exit 1
    fi
    
    # Check file size (should be > 1MB for a Go binary)
    FILE_SIZE=$(stat -f%z "$APP_DIR/bin/$SERVICE_NAME" 2>/dev/null || stat -c%s "$APP_DIR/bin/$SERVICE_NAME" 2>/dev/null || echo "0")
    if [ "$FILE_SIZE" -lt 1000000 ]; then
        echo "Error: Downloaded file is too small ($FILE_SIZE bytes). Expected > 1MB"
        echo "File content:"
        head -20 "$APP_DIR/bin/$SERVICE_NAME"
        rm -f "$APP_DIR/bin/$SERVICE_NAME"
        exit 1
    fi
    
    chmod +x "$APP_DIR/bin/$SERVICE_NAME"
    echo "Downloaded binary: $FILE_SIZE bytes"
    
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

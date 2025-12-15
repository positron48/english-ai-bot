#!/bin/bash

set -e

# Configuration - can be overridden by environment variables
GITHUB_REPO="${GITHUB_REPO:-positron48/universal-ai-bot}"
APP_DIR="${PWD}"  # Use current directory

echo "🚀 Setting up Universal AI Bot..."

# Create directories in current location
mkdir -p "$APP_DIR"/{bin,logs}

echo "📁 Creating directories..."

# Download necessary files from GitHub
echo "📥 Downloading configuration files..."

# Download .env.example if .env doesn't exist
if [ ! -f "$APP_DIR/.env" ]; then
    curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/master/env.example" > "$APP_DIR/.env"
    echo "✅ Created .env file from example"
    echo "📝 Edit the file: nano $APP_DIR/.env"
fi

# Load SERVICE_NAME from .env file
if [ -f "$APP_DIR/.env" ]; then
    # Source the .env file to load variables safely
    set -a  # automatically export all variables
    source "$APP_DIR/.env"
    set +a  # stop automatically exporting
fi
SERVICE_NAME="${SERVICE_NAME:-ai-bot}"

# Fallback: try to download from GitHub
curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/master/Makefile.deploy" > "$APP_DIR/Makefile" 2>/dev/null || {
	echo "❌ Could not find Makefile.deploy locally or download from GitHub"
	echo "Please ensure Makefile.deploy exists in the repository"
	exit 1
}
echo "✅ Downloaded Makefile.deploy from GitHub"

# Download deployment script
mkdir -p "$APP_DIR/scripts"
curl -s "https://raw.githubusercontent.com/$GITHUB_REPO/master/scripts/deploy.sh" > "$APP_DIR/scripts/deploy.sh"
chmod +x "$APP_DIR/scripts/deploy.sh"
echo "✅ Downloaded deployment script"

# Create user systemd service
echo "⚙️ Setting up systemd service..."
mkdir -p ~/.config/systemd/user
tee ~/.config/systemd/user/$SERVICE_NAME.service > /dev/null <<EOF
[Unit]
Description=Universal AI Bot
After=network.target

[Service]
Type=simple
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/$SERVICE_NAME
Restart=always
EnvironmentFile=$APP_DIR/.env

[Install]
WantedBy=default.target
EOF

# Enable user service
systemctl --user daemon-reload
systemctl --user enable $SERVICE_NAME

echo ""
echo "🎉 Setup completed!"
echo ""
echo "📋 Next steps:"
echo "1. Edit configuration: nano $APP_DIR/.env"
echo "2. Run deployment: make deploy"
echo "3. Check status: make status"
echo "4. View logs: make logs"
echo ""
echo "💡 Available commands:"
echo "   make deploy  - Deploy/update bot"
echo "   make status  - Check status"
echo "   make logs    - View logs"
echo "   make restart - Restart service"

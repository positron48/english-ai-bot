## Deployment

### Server Setup (One-time configuration)

Before deploying the bot, you need to set up the server environment. This is done once per server.

#### 1. Prerequisites

- Ubuntu/Debian server with SSH access
- User with sudo privileges
- Go 1.23+ installed (for building from source)

#### 2. One-Command Setup

SSH into your server, create a directory for the bot, and run the setup script:

```bash
# Create a directory for the bot
mkdir ai-bot && cd ai-bot

# Run setup script directly from GitHub (requires sudo for systemd service creation)
curl -sSL https://raw.githubusercontent.com/positron48/universal-ai-bot/master/scripts/setup.sh | bash
```

This script will:
- Create necessary directories in the current folder (`bin/`, `logs/`)
- Download all required files from GitHub (Makefile, deploy script, .env.example)
- Set up a systemd user service
- Enable the service for auto-start
- Configure user linger for systemd services

#### 3. Configure Environment

After setup, configure your bot:

```bash
# Edit the environment file
nano .env
```

Set the required variables:
```env
# Service Configuration
SERVICE_NAME=ai-bot

# GitHub Configuration (required for private repositories)
GITHUB_TOKEN=your_github_personal_access_token_here

# Telegram Bot Configuration
TELEGRAM_TOKEN=your_bot_token_here
AI_URL=https://openrouter.ai/api/v1
AI_API_KEY=your_openrouter_api_key
AI_MODEL=qwen/qwen3-coder:free
AI_PROMPT="You are a helpful AI assistant..."
```

**Note:** If you're using a private repository, you must set `GITHUB_TOKEN` with a GitHub Personal Access Token that has `repo` scope.

**Note:** You can customize the `SERVICE_NAME` to change the systemd service name. This is useful if you want to run multiple bot instances.

### Deployment Process

#### Manual Deployment

After setup and configuration, deploy the latest release:

```bash
# Deploy latest version (run from the bot directory)
make deploy

# Check status
make status

# View logs
make logs
```

#### Update Deployment

To update to a newer version:

```bash
make update
```

#### Available Commands

The deployment Makefile provides these commands:

```bash
make deploy  - Deploy/update the bot
make update  - Update to latest version  
make status  - Check service status
make logs    - View service logs
make restart - Restart service
make stop    - Stop service
make start   - Start service
make help    - Show help
```

### CI/CD with GitHub Actions

#### 1. GitHub Repository Secrets

Add the following secrets to your GitHub repository (`Settings` → `Secrets and variables` → `Actions`):

**Required Secrets for CI/CD:**
- `SERVER_HOST` - Your server IP address or domain (e.g., `192.168.1.100` or `your-domain.com`)
- `SERVER_USER` - SSH username for deployment (e.g., `deploy` or `ubuntu`)
- `SERVER_SSH_KEY` - Private SSH key for deployment (contents of your private key file)
- `SERVER_PATH` - Path to bot directory on server (e.g., `/var/www/english-bot`)

#### 2. Server SSH Key Configuration

For security, create a dedicated SSH key for deployment with restricted commands:

**On your local machine:**

```bash
# Generate a new SSH key for deployment
ssh-keygen -t ed25519 -f ~/.ssh/universal-bot -C "deploy@your-domain.com"

# Show your public key
cat ~/.ssh/universal-bot.pub
```

**On your server, restrict the SSH key:**

```bash
# Edit authorized_keys file
nano ~/.ssh/authorized_keys
```

Add the public key with command restrictions:

```bash
# Add this line to ~/.ssh/authorized_keys
command="cd /path/to/your/bot && make update",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... deploy@your-domain.com
```

This restriction ensures the SSH key can only:
- Execute `make update` in the bot directory
- Cannot perform port forwarding, X11 forwarding, or agent forwarding
- Cannot allocate a pseudo-terminal


#### 4. Environment Variables for Production

The deployment script will automatically create a `.env` file from `env.example`. You can customize the production environment by:

**Option 1: Edit the .env file directly on the server:**
```bash
nano .env
```

### Deployment Commands Reference

| Command | Description |
|---------|-------------|
| `curl -sSL .../setup.sh \| bash` | Initial server setup (run once) |
| `make deploy` | Deploy latest release |
| `make update` | Update to latest version |
| `make status` | Check service status |
| `make logs` | View service logs |
| `make restart` | Restart service |
| `make stop` | Stop service |
| `make start` | Start service |

### Service Management

The bot runs as a systemd user service. You can manage it with:

```bash
# Check status
systemctl --user status ai-bot

# Start service
systemctl --user start ai-bot

# Stop service
systemctl --user stop ai-bot

# Restart service
systemctl --user restart ai-bot

# View logs
journalctl --user -u ai-bot -f

# Enable auto-start
systemctl --user enable ai-bot

# Disable auto-start
systemctl --user disable ai-bot
```

### Troubleshooting Deployment

#### Common Issues:

1. **Service not starting:**
   ```bash
   # Check logs
   journalctl --user -u ai-bot -f
   
   # Check if .env file exists and is configured
   ls -la .env
   ```

2. **Permission issues:**
   ```bash
   # Fix ownership (run from bot directory)
   sudo chown -R $USER:$USER .
   ```

3. **SSH key not working:**
   ```bash
   # Test SSH connection
   ssh -i ~/.ssh/deploy_key user@your-server.com
   
   # Check authorized_keys format
   cat ~/.ssh/authorized_keys
   ```

4. **Binary not found:**
   ```bash
   # Check if binary exists
   ls -la bin/
   
   # Re-run deployment
   make deploy
   ```

5. **Setup script failed:**
   ```bash
   # Check if all files were downloaded
   ls -la
   
   # Re-run setup
   curl -sSL https://raw.githubusercontent.com/positron48/universal-ai-bot/master/scripts/setup.sh | bash
   ```

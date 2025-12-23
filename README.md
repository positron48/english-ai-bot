# AI Telegram Bot

A simple AI-powered Telegram bot written in Go that integrates with OpenAI-compatible API providers (like OpenRouter).

## Features

- AI-powered responses using OpenAI-compatible APIs
- Support for OpenRouter and other providers
- **Automatic Markdown to Telegram formatting** - converts AI responses to proper Telegram format
- **Vocabulary cards caching** - SQLite database for storing word definitions
- **Request history tracking** - tracks which users requested which words and when
- **Smart word detection** - automatically detects single-word queries and uses cached definitions when available
- **Spaced Repetition System (SRS)** - intelligent vocabulary training with adaptive scheduling
- **Training sessions** - interactive vocabulary card training with multiple choice questions
- **Session persistence** - training sessions survive bot restarts
- **Telegram Mini App** - web-based user cabinet and admin panel (API backend ready for Vue frontend)
- **JWT Authentication** - stateless auth with access/refresh tokens
- **Swagger UI** - interactive API documentation with auto Bearer prefix and token refresh
- Long polling and webhook support
- Structured logging with Zap
- Configuration management with Viper
- Graceful shutdown
- Health check endpoint
- Docker support

## 🚀 Quick Start

Want to get up and running in 2 minutes? Here's the fastest way:

### 1. Download Binary
```bash
# Download the latest release for your system
curl -L -o universal-ai-bot https://github.com/positron48/universal-ai-bot/releases/latest/download/universal-ai-bot-linux_amd64
chmod +x universal-ai-bot
```

Binaries for other systems are available on [the Releases page](https://github.com/positron48/universal-ai-bot/releases).

### 2. Create .env File
```bash
# Create your configuration file
cat > .env << 'EOF'
TELEGRAM_TOKEN=your_bot_token_here
AI_URL=https://openrouter.ai/api/v1
AI_API_KEY=your_openrouter_api_key
AI_MODEL=qwen/qwen3-coder:free
AI_PROMPT=You are a helpful AI assistant. Please respond to the user's message in a helpful and informative way.
DATABASE_PATH=./data/words.db
EOF
```

### 3. Run!
```bash
./universal-ai-bot
```

**That's it!** 🎉

Your bot is now running and ready to respond to messages in Telegram.

> **Need help?** Get your Telegram bot token from [@BotFather](https://t.me/botfather) and your AI API key from [OpenRouter](https://openrouter.ai/).

## Requirements

- Go 1.23+ (for building from source)
- Docker and Docker Compose (optional)
- Telegram Bot Token
- AI Provider API Key (OpenRouter, OpenAI, etc.)

> **Note:** The compiled binary is **fully static** and doesn't require SQLite or any other system dependencies. You can use the pre-built binary on any Linux system without installing additional packages.

## Installation and Setup

### 1. Cloning and Setup

```bash
git clone git@github.com:positron48/universal-ai-bot.git
cd ai-bot
make setup
```

### 2. Bot Configuration

1. Create a bot via [@BotFather](https://t.me/botfather)
2. Get the bot token
3. Get an API key from an AI provider (e.g., [OpenRouter](https://openrouter.ai/))
4. Edit the `.env` file:

```bash
# Required settings
TELEGRAM_TOKEN=your_bot_token_here
AI_URL=https://openrouter.ai/api/v1
AI_API_KEY=your_openrouter_api_key
AI_MODEL=gpt-3.5-turbo
AI_PROMPT=You are a helpful AI assistant. Please respond to the user's message in a helpful and informative way.

# Optional settings
TELEGRAM_DEBUG=false
LOG_LEVEL=info
SERVER_ADDRESS=:8184
DATABASE_PATH=./data/words.db
```

> **Note:** The database file and directory will be created automatically on first run if they don't exist.

### 3. Running

#### Local Development

```bash
# Install dependencies
make tidy

# Run in development mode
make dev

# Or build and run
make build
make run
```

#### Docker

```bash
# Build and run with Docker Compose
make docker-build
make docker-run

# View logs
make docker-logs

# Stop
make docker-stop
```

## How it Works

### Regular Messages
1. User sends a text message to the bot
2. Bot sends a typing indicator
3. Bot forwards the message to the AI provider with the configured system prompt
4. AI provider processes the request and returns a response
5. Bot converts Markdown formatting to Telegram format
6. Bot sends the formatted AI response back to the user

### Single Word Queries (Vocabulary Cards)
When a user sends a single word, the bot uses intelligent caching:

1. **Word Detection** - Bot detects if the message is a single word (after trimming punctuation)
2. **Database Lookup** - Bot checks SQLite database for cached definition
3. **Cache Hit** - If found:
   - Returns cached definition immediately (faster response)
   - Records the request in history (user_id, word, timestamp)
4. **Cache Miss** - If not found:
   - Requests definition from AI provider
   - Saves the definition to database for future use
   - Records the request in history
   - Returns the definition to user

This approach:
- ✅ Reduces API calls and costs
- ✅ Provides faster responses for previously requested words
- ✅ Maintains a searchable vocabulary database
- ✅ Tracks user activity and popular words

## Supported AI Providers

This bot works with any OpenAI-compatible API, including:

- **OpenRouter** - Access to multiple AI models through one API
- **OpenAI** - Direct OpenAI API access
- **Anthropic** - Claude models (if using OpenRouter)
- **Google** - Gemini models (if using OpenRouter)
- **Local models** - Any self-hosted OpenAI-compatible API

## Web Application (Vue SPA)

The bot includes a full-featured Vue 3 Single Page Application (SPA) accessible as a Telegram Mini App or in a regular browser. The frontend is embedded directly into the Go binary using `go:embed`, so no separate web server is needed.

### Features

**For Regular Users:**
- **Dashboard** - View due card counts and quick access to training
- **AI Chat** - Simple one-request-one-response AI chat interface
- **Vocabulary Management** - View all learned words with statistics and delete words from your learning
- **Web Training** - Full training session experience in the browser with the same UX as Telegram bot

**For Administrators:**
- **Circuit Breaker Management** - View status and reset circuit breaker
- **Training Cards Management** - Delete training cards by word, delete all cards, or view detailed data

### Access Methods

1. **Inside Telegram Mini App** - Open Mini App link (automatically authenticates via Telegram `initData`)
   - The app detects `window.Telegram.WebApp.initData` and automatically calls `/auth/telegram`
   - If Telegram auth fails, falls back to OTP login

2. **External Browser** - Direct URL access with OTP login
   - Navigate to `/app` on your domain
   - Enter Telegram username or ID
   - Receive OTP code via Telegram bot
   - Enter code to authenticate

### UI Routes

The SPA uses hash-based routing to avoid conflicts with API endpoints:
- `/app` or `/app/#/login` - Login page
- `/app/#/dashboard` - Dashboard
- `/app/#/vocab` - Vocabulary list
- `/app/#/training` - Training session
- `/app/#/chat` - AI chat
- `/app/#/admin` - Admin panel (admin only)

API endpoints remain at `/app/dashboard`, `/app/vocab`, etc. (without hash).

### Authentication

The web app uses **JWT (JSON Web Tokens)** for authentication:

- **Access tokens** - Short-lived (24h by default) for API requests
- **Refresh tokens** - Long-lived (30 days) for token renewal
- **Automatic token refresh** - Frontend automatically refreshes expired access tokens using refresh token
- **Token storage** - Tokens stored in browser localStorage

### Configuration

Add these environment variables to enable the web app:

```env
WEBAPP_JWT_SECRET=your-secret-key-for-jwt-signing  # Required: openssl rand -hex 32
WEBAPP_JWT_TTL_HOURS=24                            # Access token TTL (default: 24h)
WEBAPP_REFRESH_TTL_HOURS=720                       # Refresh token TTL (default: 30 days)
WEBAPP_OTP_TTL_SECONDS=300                         # OTP code expiration
WEBAPP_PUBLIC_URL=https://your-domain.com          # Optional: for CORS
```

The web app is automatically available at `/app` route on the same server. The frontend is embedded in the Go binary, so no separate build step is needed on the server - just deploy the binary.

### Building the Frontend

The frontend is built automatically during CI/CD:
1. GitHub Actions installs Node.js and builds the Vue app (`npm install && npm run build`)
2. The built `webapp/dist` folder is embedded into the Go binary using `go:embed`
3. The Go binary serves the static files at runtime

For local development:
```bash
cd webapp
npm install
npm run dev  # Runs Vite dev server (proxy to API at :8184)
```

### API Documentation

**Swagger UI** is available at `/swagger/` with:
- Auto Bearer prefix - just enter token without "Bearer "
- Auto token refresh - automatically refreshes expired access tokens
- Full API documentation for all endpoints

## Commands

The bot supports the following commands:

### User Commands

- `/start` - Start the bot and get welcome message
- `/help` - Show help message with available commands
- `/train` - Start a training session with vocabulary cards
- `/get_id` - Get your Telegram user ID

### Admin Commands

The following commands are available only to the configured admin user:

- `/reset_circuit` - Reset the circuit breaker for the training worker
- `/delete_train [word]` - Delete all training cards for a specific word
- `/delete_train_all` - Delete all training cards (cascades to user_cards and review_events)
- `/get_train_data [word]` - Get detailed information about all training cards for a word

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TELEGRAM_TOKEN` | Telegram bot token | **Required** |
| `AI_URL` | AI provider API URL | **Required** |
| `AI_API_KEY` | AI provider API key | **Required** |
| `AI_MODEL` | AI model to use | `gpt-3.5-turbo` |
| `AI_PROMPT` | System prompt for AI | **Required** (or use `AI_PROMPT_FILE`) |
| `AI_PROMPT_FILE` | Path to file containing system prompt | Alternative to `AI_PROMPT` |
| `DATABASE_PATH` | Path to SQLite database file | `./data/words.db` |
| `TELEGRAM_DEBUG` | Debug mode | `false` |
| `TELEGRAM_UPDATES_TIMEOUT` | Updates timeout | `30` |
| `TELEGRAM_WEBHOOK_ENABLE` | Enable webhook | `false` |
| `TELEGRAM_WEBHOOK_DOMAIN` | Webhook domain | - |
| `TELEGRAM_WEBHOOK_PATH` | Webhook path | `/webhook` |
| `SERVER_ADDRESS` | Server address | `:8184` |
| `LOG_LEVEL` | Logging level | `info` |
| `WEBAPP_JWT_SECRET` | Secret key for JWT signing | **Required** |
| `WEBAPP_JWT_TTL_HOURS` | Access token TTL | `24` |
| `WEBAPP_REFRESH_TTL_HOURS` | Refresh token TTL | `720` |
| `WEBAPP_OTP_TTL_SECONDS` | OTP code expiration time | `300` |
| `WEBAPP_PUBLIC_URL` | Public URL for Mini App (HTTPS) | - |
| `ADMIN_TELEGRAM_ID` | Telegram ID of admin user | `0` |

### Example Configuration for OpenRouter

```env
TELEGRAM_TOKEN=your_telegram_bot_token
AI_URL=https://openrouter.ai/api/v1
AI_API_KEY=your_openrouter_api_key
AI_MODEL=anthropic/claude-3-sonnet
AI_PROMPT=You are a helpful AI assistant. Please respond to the user's message in a helpful and informative way.
```

### Customizing AI Behavior

You can customize the AI behavior in two ways:

#### Option 1: Direct prompt in .env file
Modify the `AI_PROMPT` environment variable to change how the AI responds. You can use multi-line prompts in several ways:

**Single-line prompt:**
```env
AI_PROMPT=You are a friendly customer support assistant. Always be polite and helpful.
```

**Multi-line prompt with escaped newlines:**
```env
AI_PROMPT="You are a helpful AI assistant. You should:\n- Always be polite and respectful\n- Provide accurate information\n- Ask clarifying questions when needed\n- Keep responses concise but informative\n\nPlease respond to the user's message following these guidelines."
```

**Multi-line prompt with single quotes (works in some systems):**
```env
AI_PROMPT='You are a helpful AI assistant. You should:
- Always be polite and respectful
- Provide accurate information
- Ask clarifying questions when needed
- Keep responses concise but informative

Please respond to the user's message following these guidelines.'
```

**Using heredoc in shell scripts:**
```bash
export AI_PROMPT=$(cat <<EOF
You are a helpful AI assistant with the following characteristics:

1. Personality:
   - Friendly and approachable
   - Professional but not formal
   - Patient and understanding

2. Capabilities:
   - Answer questions accurately
   - Provide step-by-step explanations
   - Offer helpful suggestions

3. Guidelines:
   - Always be respectful
   - Admit when you don't know something
   - Ask clarifying questions when needed

Please respond to the user's message following these guidelines.
EOF
)
```

#### Option 2: Load prompt from file (Recommended for long prompts)
Instead of putting the prompt directly in the `.env` file, you can create a separate text file and reference it:

```env
AI_PROMPT_FILE=prompts/english-teacher.txt
```

**Available prompt files:**
- `prompts/simple-assistant.txt` - Basic helpful assistant
- `prompts/customer-support.txt` - Customer support specialist  
- `prompts/english-teacher.txt` - English teacher and translator

**Creating your own prompt file:**
1. Create a new file in the `prompts/` directory
2. Write your prompt in plain text (no need for escaping)
3. Reference it in `.env` with `AI_PROMPT_FILE=prompts/your-file.txt`

Example prompt file (`prompts/my-custom-prompt.txt`):
```
You are a specialized AI assistant for [your domain].

Your main responsibilities:
- Provide accurate information about [topic]
- Help users with [specific tasks]
- Maintain a [tone/style] communication style

Guidelines:
- Always be [specific behavior]
- Never [specific restrictions]
- When in doubt, [fallback behavior]

Remember: [key principles]
```

## Markdown Formatting Support

The bot automatically converts AI responses from Markdown to Telegram format. Supported formatting includes:

### ✅ **Supported Elements:**
- **Headers** (`#`, `##`, `###`) → **Bold text**
- **Bold text** (`**text**`, `__text__`) → **Bold text**
- **Italic text** (`*text*`) → _Italic text_
- **Code blocks** (```language → ```)
- **Inline code** (`` `code` ``)
- **Unordered lists** (`-`, `*`) → • Bullet points
- **Ordered lists** (`1.`, `2.`) → Numbered lists
- **Links** (`[text](url)`) → [text](url)

### 📝 **Example Conversion:**

**Input (AI Response):**
```markdown
# Welcome to AI Bot

This is **bold** and *italic* text.

## Features
- Feature 1
- Feature 2

Use `code` for examples.
```

**Output (Telegram):**
```
**Welcome to AI Bot**

This is **bold** and _italic_ text.

**Features**
• Feature 1
• Feature 2

Use `code` for examples.
```

## Project Structure

```
english-bot/
├── cmd/bot/                 # Application entry point
├── internal/
│   ├── ai/                  # AI service for provider integration
│   ├── bot/                 # Bot logic and handlers
│   ├── config/              # Configuration management
│   ├── database/            # Database initialization and migrations
│   ├── logger/              # Logging configuration
│   ├── models/              # Data models (word cards, history)
│   ├── repository/          # Database repositories
│   ├── service/             # Business logic services
│   ├── utils/               # Utility functions (Markdown conversion)
│   └── web/                 # Web application (Mini App)
│       ├── templates/       # HTML templates (embedded)
│       └── static/          # CSS and static assets (embedded)
├── prompts/                 # AI prompt files
│   ├── simple-assistant.txt
│   ├── customer-support.txt
│   └── english-teacher.txt
├── data/                    # Database files (created automatically)
│   └── words.db            # SQLite database
├── .github/workflows/       # GitHub Actions
├── Dockerfile               # Docker image
├── docker-compose.yml       # Docker Compose
├── Makefile                 # Development commands
└── README.md               # Documentation
```

## Commands

### Main Commands

```bash
make setup          # Initial project setup
make build          # Build application
make run            # Run application
make dev            # Run in development mode
make test           # Run tests
make lint           # Code linting
make clean          # Clean build artifacts
```

### Docker Commands

```bash
make docker-build   # Build Docker image
make docker-run     # Run with docker-compose
make docker-stop    # Stop containers
make docker-logs    # View logs
make docker-clean   # Clean Docker resources
make deploy         # Deploy with Docker
```

## Docker

### Building Image

```bash
docker build -t ai-telegram-bot .
```

### Running Container

```bash
docker run -d \
  --name ai-telegram-bot \
  -p 8184:8184 \
  -v $(pwd)/data:/app/data \
  -e TELEGRAM_TOKEN=your_token_here \
  -e AI_URL=https://openrouter.ai/api/v1 \
  -e AI_API_KEY=your_api_key \
  -e AI_MODEL=gpt-3.5-turbo \
  -e AI_PROMPT="You are a helpful assistant" \
  -e DATABASE_PATH=/app/data/words.db \
  ai-telegram-bot
```

> **Note:** The `-v $(pwd)/data:/app/data` volume mount persists the database between container restarts.

### Docker Compose

```bash
# Start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## API Endpoints

When running, the bot exposes the following HTTP endpoints:

**Public:**
- `GET /health` - Health check endpoint
- `GET /swagger/` - Swagger UI documentation
- `POST /auth/telegram` - Telegram initData authentication (returns JWT tokens)
- `POST /auth/telegram_unsafe` - Unsafe Telegram auth (fallback)
- `POST /auth/request_otp` - Request OTP for external login
- `POST /auth/otp` - Verify OTP and get JWT tokens
- `POST /auth/refresh` - Refresh access token using refresh token

**Protected (require JWT Bearer token):**
- `GET /app/dashboard` - User dashboard
- `GET /app/vocab` - Vocabulary list
- `POST /app/training/start` - Start training session
- `GET /app/admin` - Admin panel (requires admin access)

See `/swagger/` for full API documentation.

## Database

The bot uses SQLite to store vocabulary cards and request history. The database is automatically created on first run.

### Database Schema

**word_cards** - Stores vocabulary card definitions:
- `id` - Primary key
- `word` - Word (normalized, unique)
- `definition` - AI-generated definition
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

**word_request_history** - Tracks user requests:
- `id` - Primary key
- `user_id` - Telegram user ID
- `word` - Requested word
- `requested_at` - Request timestamp

**web_otps** - One-time passwords for external login:
- `id` - Primary key
- `user_id` - User ID
- `code_hash` - Hashed OTP code
- `expires_at` - OTP expiration time
- `consumed_at` - When OTP was used (NULL if unused)
- `created_at` - Creation timestamp

**users** - Extended with `telegram_username` field for OTP login

> **Note:** The project uses JWT tokens for authentication. Sessions are stateless and stored in tokens, not in the database.

### Database Location

By default, the database is stored at `./data/words.db`. You can change this with the `DATABASE_PATH` environment variable.

### Backup

To backup your database:
```bash
# Copy the database file
cp ./data/words.db ./data/words.db.backup

# Or use sqlite3 backup command
sqlite3 ./data/words.db ".backup './data/words.db.backup'"
```

### Viewing Database

You can inspect the database using SQLite CLI:
```bash
sqlite3 ./data/words.db

# View all words
SELECT * FROM word_cards;

# View request history
SELECT * FROM word_request_history ORDER BY requested_at DESC LIMIT 10;

# Count words per user
SELECT user_id, COUNT(*) as word_count 
FROM word_request_history 
GROUP BY user_id 
ORDER BY word_count DESC;
```

## Development

### Adding New Commands

1. Edit `internal/bot/handler.go`
2. Add a new case to the `handleCommand` function
3. Implement the command logic

### Adding New Features

1. Create new packages in `internal/`
2. Add configuration if needed
3. Update the main application

## Testing

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Check coverage
go test -cover ./...
```

## Monitoring

The bot provides a health check endpoint:

```bash
curl http://localhost:8184/health
```

## Security

- Uses non-root user in Docker
- Security scanning with Gosec in CI
- Configuration validation on startup
- API keys are loaded from environment variables

## Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is distributed under the most permissive license - the MIT License. This allows maximum freedom for use, modification, and distribution. See the `LICENSE` file for details.

## Deployment

See [deployment.md](DEPLOYMENT.md)

## Support

If you have questions or issues:

1. Check [Issues](https://github.com/your-repo/issues)
2. Create a new Issue with detailed description
3. Make sure to provide logs and configuration
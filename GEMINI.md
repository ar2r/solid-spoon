# Project Context: solid-spoon

## Overview
**solid-spoon** is a Telegram bot written in Go designed to download YouTube videos. It allows users to send YouTube links (including Shorts), select their preferred video quality, and receive the video directly in the chat. The bot handles video downloading, optional compression using `ffmpeg` for files exceeding Telegram's size limits, and user management via SQLite.

## Key Features
*   **YouTube Downloading:** Supports `youtube.com/watch`, `youtu.be`, and YouTube Shorts.
*   **Quality Selection:** Interactive inline buttons for users to choose video resolution (360p, 480p, 720p, 1080p).
*   **Smart Compression:** Automatically compresses videos larger than 50MB using `ffmpeg` to ensure they can be sent via the standard Telegram Bot API.
*   **User Management:** Stores user information in a local SQLite database.
*   **Deployment:** Dockerized for easy deployment, with CI/CD pipelines via GitHub Actions.

## Tech Stack
*   **Language:** Go 1.25
*   **Bot Framework:** `github.com/go-telegram-bot-api/telegram-bot-api/v5`
*   **YouTube Library:** `github.com/kkdai/youtube/v2`
*   **Database:** SQLite (`modernc.org/sqlite`)
*   **External Tools:** `ffmpeg` (required for video compression)

## Project Structure
```
/
├── cmd/
│   └── bot/           # Application entry point (main.go)
├── internal/
│   ├── bot/           # Bot initialization and wrapper
│   ├── database/      # SQLite connection and migrations
│   │   ├── models/    # Data models (User, etc.)
│   │   └── repository/# Data access layer
│   ├── downloader/    # YouTube download and ffmpeg compression logic
│   └── handler/       # Telegram update handlers
│       ├── start.go   # /start command handler
│       └── youtube.go # YouTube link and callback handler
├── Dockerfile         # Docker build configuration
├── go.mod             # Go module definition
└── .github/workflows/ # CI/CD workflows
```

## Setup & Running

### Prerequisites
*   Go 1.25+
*   Docker (optional, for containerized run)
*   `ffmpeg` (installed on the host system if running locally)

### Environment Variables
| Variable | Description | Required |
| :--- | :--- | :--- |
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather | Yes |
| `ADMIN_CHAT_ID` | Chat ID for admin notifications | No |
| `APP_VERSION` | Application version (injected during build) | No |

### Local Development
1.  **Clone the repository:**
    ```bash
    git clone <repo_url>
    cd solid-spoon
    ```
2.  **Set environment variables:**
    ```bash
    export TELEGRAM_BOT_TOKEN="your_token_here"
    ```
3.  **Run the bot:**
    ```bash
    go run ./cmd/bot
    ```

### Docker
1.  **Build the image:**
    ```bash
    docker build -t solid-spoon .
    ```
2.  **Run the container:**
    ```bash
    docker run -e TELEGRAM_BOT_TOKEN="your_token" solid-spoon
    ```

## Development Conventions
*   **Package Layout:** Follows the standard Go project layout (`cmd/`, `internal/`).
*   **Error Handling:** Extensive logging of errors with context.
*   **Database:** Uses `modernc.org/sqlite` (CGO-free SQLite). Ensure the `data` directory exists or is writable if persisting data.
*   **Testing:** Run tests using `go test ./...`.

## Architecture Highlights
*   **Handler Logic:** The `YouTubeHandler` (`internal/handler/youtube.go`) orchestrates the flow:
    1.  Detects YouTube link.
    2.  Fetches available formats.
    3.  Presents inline keyboard options.
    4.  Handles callback (user selection).
    5.  Downloads and optionally compresses video.
    6.  Uploads as a document.
*   **Compression:** The bot aims to stay under the 50MB limit of the standard Telegram Bot API. If a downloaded video exceeds this, it attempts to compress it using `ffmpeg`.

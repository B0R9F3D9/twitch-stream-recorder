# Twitch Stream Recorder

A lightweight Go application that automatically records Twitch streams when they go live.

## Requirements

- **Twitch Developer App Credentials** - Get `CLIENT_ID` and `CLIENT_SECRET` from [Twitch Developers Console](https://dev.twitch.tv/console/apps)
- **Twitch Authentication Token** - `WEB_API_TOKEN` from [Streamlink Authentication Guide](https://streamlink.github.io/cli/plugins/twitch.html#authentication)
- **Runtime Environment** - [Golang](https://go.dev/dl/) **or** [Docker](https://www.docker.com/)

## ⚙️ Configuration

1. Rename `.env.example` to `.env`
2. Configure your environment variables:

```env
# Twitch API credentials
CLIENT_ID=your_client_id_here
CLIENT_SECRET=your_client_secret_here
WEB_API_TOKEN=your_web_api_token_here

# Stream configuration
USERNAME=ishowspeed     # Twitch channel to monitor
QUALITY=best            # Stream quality (best, 720p, 480p, etc.)
REFRESH=15s             # Check interval (15s, 30s, 1m)

# Path configuration (required for non-Docker setup)
TEMP_PATH=./temp        # Temporary download directory
FINAL_PATH=./final      # Final recordings storage
```

3. **For non-Docker execution only**: Create the directories specified in `TEMP_PATH` and `FINAL_PATH`

## 🚀 Quick Start

### Docker Compose (Recommended)

```bash
docker-compose up --build
```

### Manual Docker Setup

1. Build the image:
```bash
docker build -t twitch-recorder .
```

2. Run the container:
```bash
docker run -d \
  --name twitch-stream-recorder \
  --env-file .env \
  -v /path/to/your/temp:/app/temp \
  -v /path/to/your/recordings:/app/final \
  twitch-recorder
```

### Native Go Execution

```bash
go mod download
go run .
```

### Volume Mapping Explanation
- `/app/temp` → Temporary files during recording (mapped to host directory)
- `/app/final` → Completed recordings storage (mapped to host directory)

**Example host paths:**
- Linux/Mac: `-v ~/twitch/temp:/app/temp -v ~/twitch/recordings:/app/final`
- Windows: `-v C:\twitch\temp:/app/temp -v C:\twitch\recordings:/app/final`

---

*Based on [ancalentari/twitch-stream-recorder](https://github.com/ancalentari/twitch-stream-recorder)*

# LocalLaunch

Open-source, cross-platform local process launcher with a secure localhost HTTP API.

LocalLaunch lets trusted web applications start arbitrary locally installed applications through a simple HTTP API running on `127.0.0.1`.

## How It Works

```
Browser Web App
    |
    | HTTP localhost API
    |
LocalLaunch
    |
    | os process execution
    |
Any local executable
```

LocalLaunch does not know about applications. It receives an executable path, arguments, and optional working directory, then starts the process.

## Install

### Recommended (curl install script)

macOS/Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/rhymn/locallaunch/main/scripts/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/rhymn/locallaunch/main/scripts/install.sh | VERSION=0.1.0 bash
```

The installer:

- Downloads the release binary for your OS/architecture
- Installs it to a user-local location
- Registers LocalLaunch as a background service that starts automatically after restart/login

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/rhymn/locallaunch/main/scripts/install.ps1 | iex
```

### From source

```bash
git clone https://github.com/rhymn/locallaunch.git
cd locallaunch
go build ./cmd/locallaunch
```

### Cross-compile

```bash
# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o locallaunch ./cmd/locallaunch

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o locallaunch ./cmd/locallaunch

# Windows
GOOS=windows GOARCH=amd64 go build -o locallaunch.exe ./cmd/locallaunch

# Linux
GOOS=linux GOARCH=amd64 go build -o locallaunch ./cmd/locallaunch
```

### Install scripts

See `scripts/install.sh` (macOS/Linux) and `scripts/install.ps1` (Windows).

## Usage

```bash
# Start the server
./locallaunch

# Show the auth token
./locallaunch token

# Show version
./locallaunch version
```

On first run, LocalLaunch creates a config file with a generated auth token:

- macOS: `~/Library/Application Support/locallaunch/config.json`
- Linux: `~/.config/locallaunch/config.json`
- Windows: `%APPDATA%\locallaunch\config.json`

## API

Base URL: `http://127.0.0.1:38471`

### GET /api/v1/status

No authentication required.

```bash
curl http://127.0.0.1:38471/api/v1/status
```

```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

### POST /api/v1/process

Requires `Authorization: Bearer <token>` header.

```bash
# Get your token
./locallaunch token

# Launch an application
curl -X POST http://127.0.0.1:38471/api/v1/process \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"/Applications/Safari.app","args":[]}'
```

Request body:

```json
{
  "path": "/path/to/application",
  "args": ["--example", "value"],
  "cwd": "/optional/working/directory"
}
```

Success response:

```json
{
  "started": true,
  "pid": 12345
}
```

Error response:

```json
{
  "started": false,
  "error": "message"
}
```

## Security

- Listens only on `127.0.0.1`
- Bearer token authentication required for process execution
- Token generated using `crypto/rand` (32 bytes, 256-bit entropy)
- Config file protected with `0600` permissions (Unix)
- No shell execution - processes are started directly via `os/exec`

## Configuration

```json
{
  "address": "127.0.0.1:38471",
  "token": "generated-token"
}
```

## Testing

```bash
go test ./...
```

## License

MIT

# go-reload

A powerful hot-reload development tool for Go applications with automatic browser reloading. Monitors your project for file changes and automatically rebuilds your application and refreshes your browser.

## Features

- **Automatic File Watching & Reload**: Monitors your project and rebuilds when files change
- **Browser Auto-Reload**: Automatically refreshes your browser when changes are detected (via WebSocket proxy)
- **Debounced Rebuilds**: Intelligently debounces rapid file changes to prevent excessive rebuilds (default 300ms window)
- **Reverse Proxy**: Built-in HTTP proxy that injects live-reload scripts into HTML responses
- **Configurable**: Support for multiple file extensions and custom build/run commands
- **Smart Directory Exclusion**: Excludes specified directories from watching
- **Clean Process Management**: Ensures proper shutdown of child processes
- **Colorful CLI Output**: Clear, color-coded terminal feedback

## Installation

```bash
go install github.com/DevJHansen/go-reload@latest
```

Or clone and build locally:

```bash
git clone https://github.com/DevJHansen/go-reload.git
cd go-reload
go build -o go-reload
```

## Usage

Run `go-reload` from anywhere inside your Go project:

```bash
go-reload
```

The tool will:

1. Find your project root (by locating `go.mod`)
2. Create a `.go-reload.yaml` config file with defaults if one doesn't exist
3. Start watching for file changes
4. Rebuild and restart your application when changes are detected

## Configuration

On first run, a `.go-reload.yaml` file is created with default settings:

```yaml
build_cmd: 'go build -o tmp/main .'
run_cmd: './tmp/main'
app_port: 3000
proxy_port: 3001
watch:
  - .go
  - .html
  - .js
  - .ts
  - .css
  - .tmpl
excl_dirs:
  - node_modules
  - tmp
  - temp
  - .git
```

### Configuration Options

- `build_cmd`: Command to build your application (e.g., `go build -o tmp/main .`)
- `run_cmd`: Command to run your application (e.g., `./tmp/main`)
- `app_port`: Your application's port (default: 3000) - the port your Go app runs on
- `proxy_port`: Proxy port (default: 3001) - the port you access in your browser for auto-reload
- `watch`: File extensions to watch for changes (supports both backend and frontend files)
- `excl_dirs`: Directories to exclude from watching

### Using Browser Auto-Reload

To use the browser auto-reload feature:

1. Configure `app_port` to match your application's server port
2. Configure `proxy_port` to the port you want to use for development
3. Run your application with `go-reload`
4. **Access your app through the proxy port** (e.g., `http://localhost:3001` instead of `http://localhost:3000`)
5. The proxy will automatically inject WebSocket client code into your HTML pages
6. When you save changes to any watched files, your browser will automatically refresh

**Note**: The proxy only injects the reload script into responses with `Content-Type: text/html`, so it won't interfere with API responses or other content types.

## How It Works

1. **File Watching**: Uses `fsnotify` to monitor your project for changes to specified file types
2. **Debouncing**: Collects rapid file changes within a 300ms window to prevent excessive rebuilds
3. **Rebuild**: When changes settle, stops the running process and rebuilds your application
4. **Restart**: Starts the newly built binary
5. **Proxy Server**: Runs a reverse proxy that forwards requests to your app while injecting WebSocket client code
6. **Browser Reload**: Waits for your server to be ready, then broadcasts a reload message via WebSocket
7. **Auto-Refresh**: Connected browsers receive the reload signal and automatically refresh
8. **Process Management**: Ensures clean shutdown of child processes and proxy server

## Stopping

Press `Ctrl/Command+C` to gracefully stop the watcher, proxy server, and your application. The tool will:

- Close all WebSocket connections
- Shut down the proxy server
- Terminate the running application process
- Clean up resources

## Project Structure

```
go-reload/
├── main.go           # Entry point, orchestrates components
├── config/           # Configuration loading and defaults
├── builder/          # Handles building and running the Go application
├── watcher/          # File system watching with fsnotify and debouncing
└── reloader/         # Reverse proxy with WebSocket-based browser reloading
    └── socket.html   # WebSocket client script injected into HTML pages
```

## Requirements

- Go 1.24.0 or later
- A Go project with `go.mod` file

## License

MIT

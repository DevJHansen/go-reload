# go-reload

A simple hot-reload development tool for Go applications. Automatically rebuilds and restarts your Go application when files change.

## Features

- Automatic file watching and reload
- Configurable build and run commands
- Support for multiple file extensions
- Excludes specified directories from watching
- Clean process management

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

- `build_cmd`: Command to build your application
- `run_cmd`: Command to run your application
- `app_port`: Your application's port (for future proxy features)
- `proxy_port`: Proxy port (for future proxy features)
- `watch`: File extensions to watch for changes
- `excl_dirs`: Directories to exclude from watching

## How It Works

1. **File Watching**: Monitors your project for changes to specified file types
2. **Rebuild**: When a change is detected, stops the running process and rebuilds
3. **Restart**: Starts the newly built binary
4. **Process Management**: Ensures clean shutdown of child processes

## Stopping

Press `Ctrl/Command+C` to stop the watcher and your application.

## License

MIT

package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/DevJHansen/go-reload/builder"
	"github.com/DevJHansen/go-reload/config"
	"github.com/DevJHansen/go-reload/reloader"
	"github.com/DevJHansen/go-reload/watcher"
	"github.com/fatih/color"
)

func printBanner() {
	banner := `
╔═════════════════════════════════════════════════════════════════════════════╗
║                                                                             ║
║   ██████╗  ██████╗       ██████╗ ███████╗██╗      ██████╗  █████╗ ██████╗   ║
║  ██╔════╝ ██╔═══██╗      ██╔══██╗██╔════╝██║     ██╔═══██╗██╔══██╗██╔══██╗  ║
║  ██║  ███╗██║   ██║█████╗██████╔╝█████╗  ██║     ██║   ██║███████║██║  ██║  ║
║  ██║   ██║██║   ██║╚════╝██╔══██╗██╔══╝  ██║     ██║   ██║██╔══██║██║  ██║  ║
║  ╚██████╔╝╚██████╔╝      ██║  ██║███████╗███████╗╚██████╔╝██║  ██║██████╔╝  ║
║   ╚═════╝  ╚═════╝       ╚═╝  ╚═╝╚══════╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝   ║
║                                                                             ║
║                     Hot Reload Development Tool for Go                      ║
║                              by Justin Hansen                               ║
║                                                                             ║
╚═════════════════════════════════════════════════════════════════════════════╝
`
	color.Cyan(banner)
}

func main() {
	printBanner()

	dir, wkDirErr := os.Getwd()
	cfg := config.DefaultConfig()

	if wkDirErr != nil {
		color.Red("failed to get working directory: %+v", wkDirErr)
		return
	}

	for {
		_, goModErr := os.Stat(filepath.Join(dir, "go.mod"))

		if goModErr == nil {
			color.Cyan("Project root: %s", dir)

			configPath := filepath.Join(dir, ".go-reload.yaml")
			_, yamlErr := os.Stat(configPath)

			if yamlErr != nil {
				if err := config.Save(configPath, cfg); err != nil {
					color.Red("Failed to save config: %v", err)
					return
				}
				color.Green("Created .go-reload.yaml with defaults")

			} else {
				yamlCfg, err := config.Load(configPath)

				if err != nil {
					color.Red("Failed to load go-reload config")
					panic("Failed to load go-reload config")
				}

				cfg = yamlCfg
				color.Cyan("Config loaded: %v", yamlCfg)
			}

			break
		}

		dir = filepath.Dir(dir)

		if filepath.Dir(dir) == dir {
			color.Red("Error: go.mod not found")
			return
		}
	}

	builder, newBuilderErr := builder.New(cfg.RunCmd)

	if newBuilderErr != nil {
		color.Red("failed to create new builder struct: %+v", newBuilderErr)
		return
	}

	buildErr := builder.Build(cfg.BuildCmd)

	if buildErr != nil {
		color.Red("failed to build: %+v", buildErr)
		return
	}

	err := builder.Start()

	if err != nil {
		color.Red("failed to start process: %+v", err)
		return
	}

	proxy, err := reloader.Start(&cfg)

	if err != nil {
		color.Red("failed to start proxy: %+v", err)
		return
	}

	w, err := watcher.New(dir, builder, &cfg, proxy)

	if err != nil {
		color.Red("Error starting watcher: %+v", err)
		color.Yellow("Shutting down...")
		if builderErr := builder.Stop(); builderErr != nil {
			color.Red("Error stopping builder: %v", builderErr)
		}

		if proxyErr := proxy.Stop(); proxyErr != nil {
			color.Red("Error stopping proxy: %v", proxyErr)
		}
		return
	}

	go w.Watch()

	color.Green("Server started")

	defer func() {
		color.Yellow("Cleaning up...")
		if builderErr := builder.Stop(); builderErr != nil {
			color.Red("Error stopping builder: %v", builderErr)
		}

		if proxyErr := proxy.Stop(); proxyErr != nil {
			color.Red("Error stopping proxy: %v", proxyErr)
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	color.Yellow("Got the signal, shutting down...")
	builder.Stop()
	color.Green("Shut down complete")
}

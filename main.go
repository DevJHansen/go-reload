package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/DevJHansen/go-reload/builder"
	"github.com/DevJHansen/go-reload/config"
	"github.com/DevJHansen/go-reload/reloader"
	"github.com/DevJHansen/go-reload/watcher"
)

func main() {
	dir, wkDirErr := os.Getwd()   // working directly from where we run our go-reload command
	cfg := config.DefaultConfig() // create default config

	if wkDirErr != nil {
		fmt.Printf("failed to get working directory: %+v\n", wkDirErr)
		return
	}

	// starting at current working dir, walk our way up until we find the project root dir
	// We do this because it allows us to run go-reload anywhere inside our project and have it work
	for {
		_, goModErr := os.Stat(filepath.Join(dir, "go.mod"))

		if goModErr == nil { // project root found
			fmt.Println("Project root:", dir)

			configPath := filepath.Join(dir, ".go-reload.yaml") // create config file path
			_, yamlErr := os.Stat(configPath)                   // confirm it's there

			if yamlErr != nil { // the config yaml file does not exist
				if err := config.Save(configPath, cfg); err != nil { // try save the config file and handle errors
					fmt.Printf("Failed to save config: %v\n", err)
					return
				}
				fmt.Println("Created .go-reload.yaml with defaults")

			} else { // the file does exist so load it and handle any loading errors
				yamlCfg, err := config.Load(configPath)

				if err != nil {
					panic("Failed to load go-reload config")
				}

				cfg = yamlCfg // set the config to the one loaded from the file
				fmt.Printf("Config loaded: %v\n", yamlCfg)
			}

			break // break out the loop because we have either found or generated our config
		}

		dir = filepath.Dir(dir) // set dir to one level up

		if filepath.Dir(dir) == dir { // we cant move up anymore
			fmt.Println("Error: go.mod not found")
			return
		}
	}

	builder := builder.New(cfg.RunCmd)
	buildErr := builder.Build(cfg.BuildCmd)

	if buildErr != nil {
		fmt.Printf("failed to build: %+v\n", buildErr)
		return
	}

	err := builder.Start()

	if err != nil {
		fmt.Printf("failed to start process: %+v\n", err)
		return
	}

	proxy, err := reloader.Start(&cfg)

	if err != nil {
		fmt.Printf("failed to start proxy: %+v\n", err)
		return
	}

	w, err := watcher.New(dir, builder, &cfg, proxy)

	if err != nil {
		fmt.Printf("Error starting watcher: %+v", err)
		fmt.Println("Shutting down...")
		if builderErr := builder.Stop(); builderErr != nil {
			fmt.Printf("Error stopping builder: %v\n", builderErr)
		}

		if proxyErr := proxy.Stop(); proxyErr != nil {
			fmt.Printf("Error stopping proxy: %v\n", proxyErr)
		}
		return
	}

	go w.Watch()

	fmt.Println("Server started")

	// Ensure we always try to stop on exit
	defer func() {
		fmt.Println("Cleaning up...")
		if builderErr := builder.Stop(); builderErr != nil {
			fmt.Printf("Error stopping builder: %v\n", builderErr)
		}

		if proxyErr := proxy.Stop(); proxyErr != nil {
			fmt.Printf("Error stopping proxy: %v\n", proxyErr)
		}
	}()

	// 1. Create mailbox
	sigChan := make(chan os.Signal, 1)

	// 2. Tell OS: "Put (Ctrl/Command)+C messages in this mailbox"
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 3. Wait at the mailbox until a message arrives
	<-sigChan // Program pauses here

	// 4. Message received! Continue...
	fmt.Println("Got the signal, shutting down...")
	builder.Stop()
	fmt.Println("Shut down complete")
}

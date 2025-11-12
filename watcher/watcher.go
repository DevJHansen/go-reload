package watcher

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"slices"

	"github.com/DevJHansen/go-reload/builder"
	"github.com/DevJHansen/go-reload/config"
	"github.com/DevJHansen/go-reload/reloader"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	projectRoot string
	builder     *builder.Builder
	config      *config.Config
	watcher     *fsnotify.Watcher
	proxy       *reloader.Proxy
}

func New(projectRoot string, b *builder.Builder, c *config.Config, p *reloader.Proxy) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	w := &Watcher{
		projectRoot: projectRoot,
		builder:     b,
		config:      c,
		watcher:     fsWatcher,
		proxy:       p,
	}

	if err := w.addDirectories(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addDirectories() error {
	// Starting at our project root recursively walk each directory and add it if it's not an excluded dir
	return filepath.Walk(w.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		dirName := filepath.Base(path)

		isExclDir := slices.Contains(w.config.ExclDirs, dirName)

		if strings.HasPrefix(dirName, ".") || isExclDir {
			return filepath.SkipDir
		}

		return w.watcher.Add(path)
	})
}

func (w *Watcher) waitForServer() {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", w.config.AppPort), time.Second)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(time.Second)
	}
}

func (w *Watcher) Watch() {
	var debounceTimer *time.Timer
	debounceDuration := 300 * time.Millisecond

	for {
		select {
		case event := <-w.watcher.Events:
			isRestartEvent := event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Rename) || event.Op.Has(fsnotify.Remove)
			ext := filepath.Ext(event.Name)
			isWatchedFile := slices.Contains(w.config.Watch, ext)

			if isRestartEvent && isWatchedFile {
				fmt.Printf("File Changed: %s\n", event.Name)

				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(debounceDuration, func() {
					// Stop the running server
					w.builder.Stop()

					// Rebuild
					fmt.Println("Rebuilding...")
					if err := w.builder.Build(w.config.BuildCmd); err != nil {
						fmt.Printf("Build failed: %v\n", err)
						return // Don't start if build failed
					}

					// Start the new version
					w.builder.Start()
					w.waitForServer()
					w.proxy.Broadcast("reload")
				})
			}

		case err := <-w.watcher.Errors:
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

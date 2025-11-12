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
	"github.com/fatih/color"
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
				color.Yellow("File Changed: %s", event.Name)

				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(debounceDuration, func() {
					w.builder.Stop()

					color.Cyan("Rebuilding...")
					if err := w.builder.Build(w.config.BuildCmd); err != nil {
						color.Red("Build failed: %v", err)
						return
					}

					w.builder.Start()
					w.waitForServer()
					w.proxy.Broadcast("reload")
				})
			}

		case err := <-w.watcher.Errors:
			color.Red("Watcher error: %v", err)
		}
	}
}

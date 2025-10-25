package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"slices"

	"github.com/DevJHansen/go-reload/builder"
	"github.com/DevJHansen/go-reload/config"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	projectRoot string
	builder     *builder.Builder
	config      *config.Config
	watcher     *fsnotify.Watcher
}

func New(projectRoot string, b *builder.Builder, c *config.Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	w := &Watcher{
		projectRoot: projectRoot,
		builder:     b,
		config:      c,
		watcher:     fsWatcher,
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

func (w *Watcher) Watch() {
	for {
		select {
		case event := <-w.watcher.Events:
			isRestartEvent := event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Rename) || event.Op.Has(fsnotify.Remove)
			ext := filepath.Ext(event.Name)
			isWatchedFile := slices.Contains(w.config.Watch, ext)

			if isRestartEvent && isWatchedFile {
				fmt.Printf("File Changed: %s\n", event.Name)

				// Stop the running server
				w.builder.Stop()

				// Rebuild
				fmt.Println("Rebuilding...")
				if err := w.builder.Build(w.config.BuildCmd); err != nil {
					fmt.Printf("Build failed: %v\n", err)
					continue // Don't start if build failed
				}

				// Start the new version
				w.builder.Start()
			}

		case err := <-w.watcher.Errors:
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

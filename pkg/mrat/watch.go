package mrat

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(ctx context.Context, path string, srv *Server) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	if err := watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			if err := srv.EvalScript(string(buf), path); err != nil {
				fmt.Println("failed to eval script:", err)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher stopped")
			}
			fmt.Println("error:", err)
		}
	}
}

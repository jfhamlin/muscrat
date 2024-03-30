package mrat

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

func watchFile(ctx context.Context, path string, srv *Server) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	if err := watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	defer watcher.Close()

	if err := srv.EvalScript(path); err != nil {
		fmt.Println("failed to eval script:", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if err := srv.EvalScript(path); err != nil {
				pubsub.Publish("console.debug", map[string]interface{}{
					"message": "failed to eval script",
					"error":   err.Error(),
				})
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

func WatchScriptFile(ctx context.Context, path string, srv *Server) error {
	return watchFile(ctx, path, srv)
}

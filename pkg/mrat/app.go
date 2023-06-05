package mrat

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type (
	App struct {
		// wails context
		ctx context.Context

		srv *Server

		// file to watch
		file        string
		watchCtx    context.Context
		watchCancel context.CancelFunc
		watchDone   chan struct{}

		mtx sync.Mutex
	}
)

func NewApp(srv *Server) *App {
	return &App{
		srv: srv,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	a.srv.Start(ctx)

	if err := a.SelectFile(); err != nil {
		fmt.Println("failed to select file:", err)
	}
}

func (a *App) SelectFile() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select a file",
		DefaultDirectory: wd,
		DefaultFilename:  "synth.glj",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Script files (*.glj)",
				Pattern:     "*.glj",
			},
		},
	})
	if err != nil {
		return err
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()
	if a.watchCtx != nil {
		a.watchCancel()
		<-a.watchDone
	}

	a.file = file
	a.watchCtx, a.watchCancel = context.WithCancel(a.ctx)
	a.watchDone = make(chan struct{})

	go func() {
		defer close(a.watchDone)
		watchFile(a.watchCtx, file, a.srv)
	}()

	return nil
}

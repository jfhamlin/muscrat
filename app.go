package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type (
	App struct {
		ctx context.Context

		srv            *mrat.Server
		cancelPlayFile func()
		playFileStop   chan struct{}

		mtx sync.Mutex
	}

	OpenFileDialogResponse struct {
		FileName string
		Content  string
	}
)

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		playFileStop: make(chan struct{}),
		srv:          mrat.NewServer(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.srv.Start(context.Background())
}

// OpenFileDialog opens a file dialog.
func (a *App) OpenFileDialog() (*OpenFileDialogResponse, error) {
	fileName, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Glojure Files (*.glj)",
				Pattern:     "*.glj",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	buf, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return &OpenFileDialogResponse{
		FileName: fileName,
		Content:  string(buf),
	}, nil
}

// SaveFile saves a file. If the fileName is empty, a file dialog is
// shown. Returns the filename and an error.
func (a *App) SaveFile(fileName string, content string) (string, error) {
	if fileName == "" {
		var err error
		fileName, err = runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title: "Save File",
			Filters: []runtime.FileFilter{
				{
					DisplayName: "Glojure Files (*.glj)",
					Pattern:     "*.glj",
				},
			},
		})
		if err != nil {
			return "", err
		}
		if fileName == "" {
			return "", fmt.Errorf("no file selected")
		}
	}

	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

// PlayFile plays a file. The file is re-evaluated whenever it
// changes.
func (a *App) PlayFile(fileName string) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.stopFile()

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelPlayFile = cancel

	go func() {
		defer func() {
			a.playFileStop <- struct{}{}
		}()
		if err := mrat.WatchScriptFile(ctx, fileName, a.srv); err != nil {
			fmt.Printf("error watching script file: %v\n", err)
			// TODO: send error to UI
			return
		}
	}()

	return nil
}

func (a *App) stopFile() {
	if a.cancelPlayFile == nil {
		return
	}
	a.cancelPlayFile()
	a.cancelPlayFile = nil
	<-a.playFileStop
}

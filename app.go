package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type (
	App struct {
		ctx context.Context
	}

	OpenFileDialogResponse struct {
		FileName string
		Content  string
	}
)

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
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

package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type (
	App struct {
		ctx context.Context

		srv            *mrat.Server
		cancelPlayFile func()
		playFileStop   chan struct{}

		channelBuffers [][]float64

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

	// send at ~15 times per second, in multiples of conf.BufferSize
	maxBuffersSamples := conf.SampleRate / 15
	// round to nearest multiple of conf.BufferSize
	maxBuffersSamples = (maxBuffersSamples/conf.BufferSize + 1) * conf.BufferSize

	pubsub.Subscribe("samples", func(event string, data any) {
		if samples, ok := data.([][]float64); ok {
			a.mtx.Lock()
			if len(a.channelBuffers) != len(samples) {
				a.channelBuffers = make([][]float64, len(samples))
			}
			for i := range samples {
				a.channelBuffers[i] = append(a.channelBuffers[i], samples[i]...)
			}
			if len(a.channelBuffers[0]) >= maxBuffersSamples {
				cpy := make([][]float64, len(a.channelBuffers))
				for i := range a.channelBuffers {
					cpy[i] = make([]float64, len(a.channelBuffers[i]))
					copy(cpy[i], a.channelBuffers[i])
				}
				go runtime.EventsEmit(ctx, "samples", cpy)
				for i := range a.channelBuffers {
					a.channelBuffers[i] = a.channelBuffers[i][:0]
				}
			}
			a.mtx.Unlock()
		}
	})

	pubsub.Subscribe(ugen.KnobsChangedEvent, func(event string, data any) {
		// send the new knobs to the UI
		go func() {
			runtime.EventsEmit(ctx, "knobs-changed", ugen.GetKnobs())
		}()
	})

	// forward knob value changes from the UI to the pubsub
	runtime.EventsOn(ctx, "knob-value-change", func(data ...any) {
		id := data[0].(float64)
		value := data[1].(float64)
		update := ugen.KnobUpdate{
			ID:    uint64(id),
			Value: value,
		}
		pubsub.Publish(ugen.KnobValueChangeEvent, update)
	})

	pubsub.Subscribe("console.log", func(event string, data any) {
		go runtime.EventsEmit(ctx, "console.log", data)
	})
}

func (a *App) GetSampleRate() int {
	return conf.SampleRate
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

func (a *App) Silence() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.stopFile()

	go a.srv.PlayGraph(mrat.ZeroGraph())
}

func (a *App) stopFile() {
	if a.cancelPlayFile == nil {
		return
	}

	a.cancelPlayFile()
	a.cancelPlayFile = nil
	<-a.playFileStop
}

func (a *App) GetNSPublics() []mrat.Symbol {
	return mrat.GetNSPublics()
}

func (a *App) GetKnobs() []*ugen.Knob {
	return ugen.GetKnobs()
}

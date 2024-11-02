package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/ugen"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type (
	MuscratService struct {
		app *application.App

		srv            *mrat.Server
		cancelPlayFile func()
		playFileStop   chan struct{}

		channelBuffers [][]float64

		hydraWindow *application.WebviewWindow
		knobsWindow *application.WebviewWindow

		windowMtx sync.Mutex

		playMtx sync.Mutex
	}

	OpenFileDialogResponse struct {
		FileName string
		Content  string
	}
)

// NewMuscratService creates a new MuscratService application struct.
func NewMuscratService() *MuscratService {
	return &MuscratService{
		playFileStop: make(chan struct{}),
		srv:          mrat.NewServer(),
	}
}

// startup is called when the app starts. The app is saved so we can
// call the runtime methods
func (a *MuscratService) startup(app *application.App) {
	a.app = app

	a.srv.Start(context.Background())

	// send at ~15 times per second, in multiples of conf.BufferSize
	maxBuffersSamples := conf.SampleRate / 15
	// round to nearest multiple of conf.BufferSize
	maxBuffersSamples = (maxBuffersSamples/conf.BufferSize + 1) * conf.BufferSize

	pubsub.Subscribe("samples", func(event string, data any) {
		samples, ok := data.([][]float64)
		if !ok {
			return
		}

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
			go app.Events.Emit(&application.WailsEvent{
				Name: "samples",
				Data: cpy,
			})

			// publish RMS and max values for each channel
			rms := make([]float64, len(a.channelBuffers))
			max := make([]float64, len(a.channelBuffers))
			for i := range a.channelBuffers {
				sum := 0.0
				maxVal := 0.0
				for _, v := range a.channelBuffers[i] {
					sum += v * v
					if v > maxVal {
						maxVal = v
					}
				}
				rms[i] = math.Sqrt(sum / float64(len(a.channelBuffers[i])))
				max[i] = maxVal
			}
			go app.Events.Emit(&application.WailsEvent{
				Name: "volume",
				Data: map[string]any{
					"rms":  rms,
					"peak": max,
				},
			})

			for i := range a.channelBuffers {
				a.channelBuffers[i] = a.channelBuffers[i][:0]
			}
		}
	})

	pubsub.Subscribe(ugen.KnobsChangedEvent, func(event string, data any) {
		a.windowMtx.Lock()
		defer a.windowMtx.Unlock()

		a.updateKnobsWindow()

		// send the new knobs to the UI
		go func() {
			app.Events.Emit(&application.WailsEvent{
				Name: "knobs-changed",
				Data: ugen.GetKnobs(),
			})
		}()
	})

	// forward knob value changes from the UI to the pubsub
	app.Events.On("knob-value-change", func(evt *application.WailsEvent) {
		data := evt.Data.([]any)
		id := data[0].(float64)
		value := data[1].(float64)
		update := ugen.KnobUpdate{
			ID:    uint64(id),
			Value: value,
		}
		pubsub.Publish(ugen.KnobValueChangeEvent, update)
	})

	pubsub.Subscribe("console.log", func(event string, data any) {
		go app.Events.Emit(&application.WailsEvent{
			Name: "console.log",
			Data: data,
		})
	})

	pubsub.Subscribe("", func(event string, data any) {
		switch event {
		case "samples", ugen.KnobsChangedEvent, "console.log", "knob-value-change":
		default:
			app.Events.Emit(&application.WailsEvent{
				Name: event,
				Data: data,
			})
		}
	})
}

func (a *MuscratService) GetSampleRate() int {
	return conf.SampleRate
}

// OpenFileDialog opens a file dialog.
func (a *MuscratService) OpenFileDialog() (*OpenFileDialogResponse, error) {
	fileName, err := application.OpenFileDialog().
		SetTitle("Open File").
		AddFilter("Glojure Files (*.glj)", "*.glj").
		PromptForSingleSelection()
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
func (a *MuscratService) SaveFile(fileName string, content string) (string, error) {
	if fileName == "" {
		var err error
		fileName, err = application.SaveFileDialog().
			AddFilter("Glojure Files (*.glj)", "*.glj").
			CanCreateDirectories(true).
			PromptForSingleSelection()
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
func (a *MuscratService) PlayFile(fileName string) error {
	a.playMtx.Lock()
	defer a.playMtx.Unlock()

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

func (a *MuscratService) Silence() {
	a.playMtx.Lock()
	defer a.playMtx.Unlock()

	a.stopFile()

	go a.srv.PlayGraph(mrat.ZeroGraph())
}

func (a *MuscratService) stopFile() {
	if a.cancelPlayFile == nil {
		return
	}

	a.cancelPlayFile()
	a.cancelPlayFile = nil
	<-a.playFileStop
}

func (a *MuscratService) GetNSPublics() []mrat.Symbol {
	return mrat.GetNSPublics()
}

func (a *MuscratService) GetKnobs() []*ugen.Knob {
	return ugen.GetKnobs()
}

func (a *MuscratService) updateKnobsWindow() {
	knobs := ugen.GetKnobs()
	if len(knobs) == 0 {
		if a.knobsWindow != nil {
			a.knobsWindow.Close()
			a.knobsWindow = nil
		}
		return
	}
	if a.knobsWindow != nil {
		return
	}

	a.knobsWindow = a.app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title: "muscrat - knobs",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/knobs",
		Width:            300,
		Height:           600,
		MinWidth:         300,
		MinHeight:        300,
	})
	a.knobsWindow.On(events.Common.WindowClosing, func(e *application.WindowEvent) {
		a.windowMtx.Lock()
		defer a.windowMtx.Unlock()

		a.knobsWindow = nil
	})
}

func (a *MuscratService) ToggleHydraWindow() {
	a.windowMtx.Lock()
	defer a.windowMtx.Unlock()

	if a.hydraWindow != nil {
		return
	}

	a.hydraWindow = a.app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title: "muscrat - Hydra",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/hydra",
	})
	a.hydraWindow.On(events.Common.WindowClosing, func(e *application.WindowEvent) {
		a.windowMtx.Lock()
		defer a.windowMtx.Unlock()

		a.hydraWindow = nil
	})
}

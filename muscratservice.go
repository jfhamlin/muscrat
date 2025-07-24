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

		windowMtx sync.Mutex

		playMtx sync.Mutex

		// Volume metering state
		volumeMeterMtx      sync.Mutex
		fastRMSBuffers      [][]float64
		slowRMSBuffers      [][]float64
		currentRMS          []float64
		currentPeak         []float64
		smoothedRMS         []float64
		smoothedPeak        []float64
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

	a.srv.Start(context.Background(), false)

	// send at ~15 times per second, in multiples of conf.BufferSize
	maxBuffersSamples := conf.SampleRate / 15
	// round to nearest multiple of conf.BufferSize
	maxBuffersSamples = (maxBuffersSamples/conf.BufferSize + 1) * conf.BufferSize

	// Time constants for dual RMS windows
	fastWindowSamples := int(float64(conf.SampleRate) * 0.020) // 20ms fast window
	slowWindowSamples := int(float64(conf.SampleRate) * 0.300) // 300ms slow window

	// Ballistics constants (at ~15Hz update rate)
	// Attack: 0.95 means ~3 updates to reach 95% (200ms at 15Hz)
	// Release: 0.25 means ~12 updates to decay to 5% (800ms at 15Hz)
	attackRate := 0.95  // Fast attack
	releaseRate := 0.25 // Moderate release for better responsiveness

	pubsub.Subscribe("samples", func(event string, data any) {
		samples, ok := data.([][]float64)
		if !ok {
			return
		}

		a.volumeMeterMtx.Lock()
		defer a.volumeMeterMtx.Unlock()

		// Initialize buffers if needed
		if len(a.channelBuffers) != len(samples) {
			a.channelBuffers = make([][]float64, len(samples))
			a.fastRMSBuffers = make([][]float64, len(samples))
			a.slowRMSBuffers = make([][]float64, len(samples))
			a.currentRMS = make([]float64, len(samples))
			a.currentPeak = make([]float64, len(samples))
			a.smoothedRMS = make([]float64, len(samples))
			a.smoothedPeak = make([]float64, len(samples))
		}

		// Append new samples to all buffers
		for i := range samples {
			a.channelBuffers[i] = append(a.channelBuffers[i], samples[i]...)
			a.fastRMSBuffers[i] = append(a.fastRMSBuffers[i], samples[i]...)
			a.slowRMSBuffers[i] = append(a.slowRMSBuffers[i], samples[i]...)

			// Trim fast buffer to window size
			if len(a.fastRMSBuffers[i]) > fastWindowSamples {
				a.fastRMSBuffers[i] = a.fastRMSBuffers[i][len(a.fastRMSBuffers[i])-fastWindowSamples:]
			}

			// Trim slow buffer to window size
			if len(a.slowRMSBuffers[i]) > slowWindowSamples {
				a.slowRMSBuffers[i] = a.slowRMSBuffers[i][len(a.slowRMSBuffers[i])-slowWindowSamples:]
			}
		}

		// Process when we have enough samples
		if len(a.channelBuffers[0]) >= maxBuffersSamples {
			cpy := make([][]float64, len(a.channelBuffers))
			for i := range a.channelBuffers {
				cpy[i] = make([]float64, len(a.channelBuffers[i]))
				copy(cpy[i], a.channelBuffers[i])
			}
			go app.EmitEvent("samples", cpy)

			// Calculate RMS and peak values with improvements
			rms := make([]float64, len(a.channelBuffers))
			peak := make([]float64, len(a.channelBuffers))
			rmsDB := make([]float64, len(a.channelBuffers))
			peakDB := make([]float64, len(a.channelBuffers))

			for i := range a.channelBuffers {
				// Calculate fast RMS
				fastSum := 0.0
				for _, v := range a.fastRMSBuffers[i] {
					fastSum += v * v
				}
				fastRMS := math.Sqrt(fastSum / float64(len(a.fastRMSBuffers[i])))

				// Calculate slow RMS
				slowSum := 0.0
				for _, v := range a.slowRMSBuffers[i] {
					slowSum += v * v
				}
				slowRMS := math.Sqrt(slowSum / float64(len(a.slowRMSBuffers[i])))

				// Use the maximum of fast and slow RMS
				a.currentRMS[i] = math.Max(fastRMS, slowRMS)

				// Find peak in current buffer
				maxVal := 0.0
				for _, v := range a.channelBuffers[i] {
					absV := math.Abs(v)
					if absV > maxVal {
						maxVal = absV
					}
				}
				a.currentPeak[i] = maxVal

				// Apply ballistics
				if a.currentRMS[i] > a.smoothedRMS[i] {
					// Fast attack
					a.smoothedRMS[i] += (a.currentRMS[i] - a.smoothedRMS[i]) * attackRate
				} else {
					// Slow release
					a.smoothedRMS[i] += (a.currentRMS[i] - a.smoothedRMS[i]) * releaseRate
				}

				if a.currentPeak[i] > a.smoothedPeak[i] {
					// Instant attack for peaks
					a.smoothedPeak[i] = a.currentPeak[i]
				} else {
					// Slow decay for peaks
					a.smoothedPeak[i] += (a.currentPeak[i] - a.smoothedPeak[i]) * releaseRate
				}

				// Store linear values for compatibility
				rms[i] = a.smoothedRMS[i]
				peak[i] = a.smoothedPeak[i]

				// Convert to dB (with -60dB floor)
				const minDB = -60.0
				if a.smoothedRMS[i] > 0 {
					rmsDB[i] = 20 * math.Log10(a.smoothedRMS[i])
					rmsDB[i] = math.Max(rmsDB[i], minDB)
				} else {
					rmsDB[i] = minDB
				}

				if a.smoothedPeak[i] > 0 {
					peakDB[i] = 20 * math.Log10(a.smoothedPeak[i])
					peakDB[i] = math.Max(peakDB[i], minDB)
				} else {
					peakDB[i] = minDB
				}
			}

			go app.EmitEvent("volume", map[string]any{
				"rms":    rms,
				"peak":   peak,
				"rmsDB":  rmsDB,
				"peakDB": peakDB,
			})

			// Clear main buffer
			for i := range a.channelBuffers {
				a.channelBuffers[i] = a.channelBuffers[i][:0]
			}
		}
	})

	pubsub.Subscribe(ugen.KnobsChangedEvent, func(event string, data any) {
		a.windowMtx.Lock()
		defer a.windowMtx.Unlock()

		// send the new knobs to the UI
		go func() {
			app.EmitEvent("knobs-changed", ugen.GetKnobs())
		}()
	})

	// forward knob value changes from the UI to the pubsub
	app.OnEvent("knob-value-change", func(evt *application.CustomEvent) {
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
		go app.EmitEvent("console.log", data)
	})

	// Subscribe to scope events
	pubsub.Subscribe("scope.data", func(event string, data any) {
		go app.EmitEvent("scope.data", data)
	})

	pubsub.Subscribe("scopes-changed", func(event string, data any) {
		fmt.Println("scopes-changed event received", data)
		go app.EmitEvent("scopes-changed", data)
	})

	// Handle trigger level changes from frontend
	app.OnEvent("scope.setTriggerLevel", func(evt *application.CustomEvent) {
		data := evt.Data.([]any)
		id := data[0].(string)
		level := data[1].(float64)
		update := ugen.TriggerUpdate{
			ID:    id,
			Level: level,
		}
		pubsub.Publish(ugen.ScopeTriggerChangeEvent, update)
	})

	pubsub.Subscribe("", func(event string, data any) {
		switch event {
		case "samples", ugen.KnobsChangedEvent, "console.log", "knob-value-change", "scope.data", "scopes-changed":
		default:
			app.EmitEvent(event, data)
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

// SaveToTemp saves content to a temporary file. Returns the temp file path.
func (a *MuscratService) SaveToTemp(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "muscrat_live_*.glj")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
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
	a.hydraWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		a.windowMtx.Lock()
		defer a.windowMtx.Unlock()

		a.hydraWindow = nil
	})
}

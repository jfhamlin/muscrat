package ugen

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/jfhamlin/muscrat/pkg/webaudio"
)

type (
	// WebAudioNode is a UGen that broadcasts audio parameters to
	// connected web clients over WebSocket. It samples input signals
	// at a low frequency and sends them to mobile/web devices for
	// synthesis and visualization.
	WebAudioNode struct {
		// Server instance
		server *webaudio.Server

		// Port for the HTTP server
		Port int `json:"port"`

		// Sample rate tracking
		sampleRate int

		// Parameter names to track
		ParamNames []string `json:"params"`

		// Sampling control
		samplesPerUpdate int
		sampleCounter    atomic.Uint64

		// Last sent values (to avoid redundant sends)
		lastValues map[string]float64
		valuesMu   sync.RWMutex

		// Lifecycle
		started atomic.Bool
	}
)

// NewWebAudioNode creates a new WebAudio node that broadcasts parameters
// to connected web clients.
//
// Parameters:
//   - port: The HTTP server port (ngrok will tunnel this)
//   - updateHz: How often to sample and send parameters (e.g., 20 for 20Hz)
//   - paramNames: Names of the input parameters to track
func NewWebAudioNode(port int, updateHz float64, paramNames []string) *WebAudioNode {
	if port == 0 {
		port = 8765 // Default port
	}
	if updateHz == 0 {
		updateHz = 20 // Default to 20Hz updates
	}

	return &WebAudioNode{
		Port:       port,
		ParamNames: paramNames,
		lastValues: make(map[string]float64),
		// samplesPerUpdate will be calculated in Start() when we know sample rate
	}
}

func (w *WebAudioNode) Start(ctx context.Context) error {
	if w.started.Swap(true) {
		return nil // Already started
	}

	// Create and start server
	w.server = webaudio.NewServer(w.Port)
	if err := w.server.Start(); err != nil {
		return fmt.Errorf("failed to start WebAudio server: %w", err)
	}

	log.Printf("WebAudioNode started: %s", w.server.GetURL())
	log.Printf("Open this URL on your phone to receive audio from Muscrat")
	log.Printf("Tracking parameters: %v", w.ParamNames)

	return nil
}

func (w *WebAudioNode) Stop(ctx context.Context) error {
	if !w.started.Swap(false) {
		return nil // Already stopped
	}

	if w.server != nil {
		log.Printf("WebAudioNode stopping (had %d clients)", w.server.GetClientCount())
		return w.server.Stop()
	}

	return nil
}

func (w *WebAudioNode) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	// Calculate samples per update if not set
	if w.samplesPerUpdate == 0 && cfg.SampleRateHz > 0 {
		// Default to 20Hz updates if not specified
		w.samplesPerUpdate = cfg.SampleRateHz / 20
		w.sampleRate = cfg.SampleRateHz
		log.Printf("WebAudioNode: sampling at %d samples per update (~20Hz)", w.samplesPerUpdate)
	}

	// Fill output with zeros (this is a control node, not an audio generator)
	for i := range out {
		out[i] = 0
	}

	// Check if it's time to send an update
	currentSample := w.sampleCounter.Add(uint64(len(out)))
	if w.samplesPerUpdate == 0 || currentSample%uint64(w.samplesPerUpdate) > uint64(len(out)) {
		return // Not time yet
	}

	// Sample the last value from each input parameter
	params := make(map[string]float64)
	hasChanges := false

	for _, paramName := range w.ParamNames {
		if samples, ok := cfg.InputSamples[paramName]; ok && len(samples) > 0 {
			// Get the last sample value
			value := samples[len(samples)-1]

			// Check for NaN and replace with 0
			if value != value {
				value = 0
			}

			params[paramName] = value

			// Check if value changed significantly
			w.valuesMu.RLock()
			lastValue, exists := w.lastValues[paramName]
			w.valuesMu.RUnlock()

			if !exists || abs(value-lastValue) > 0.001 {
				hasChanges = true
			}
		}
	}

	// Only send if we have changes and clients
	if hasChanges && w.server != nil && w.server.GetClientCount() > 0 {
		w.valuesMu.Lock()
		w.lastValues = params
		w.valuesMu.Unlock()

		// Broadcast to all connected clients
		w.server.Broadcast(webaudio.Message{
			Type:   "param",
			Params: params,
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetURL returns the public ngrok URL for this node
func (w *WebAudioNode) GetURL() string {
	if w.server == nil {
		return ""
	}
	return w.server.GetURL()
}

// GetClientCount returns the number of connected clients
func (w *WebAudioNode) GetClientCount() int {
	if w.server == nil {
		return 0
	}
	return w.server.GetClientCount()
}

// SendCommand sends a control command to all clients
func (w *WebAudioNode) SendCommand(action string) {
	if w.server != nil {
		w.server.Broadcast(webaudio.Message{
			Type:   "synth",
			Action: action,
		})
	}
}

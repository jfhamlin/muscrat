package ugen

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

// Scope is a unit generator that passes audio through unchanged while
// publishing sample buffers for visualization
type Scope struct {
	id         string
	name       string
	bufferSize int

	// Circular buffer for samples
	buffer      []float64
	bufferIndex int

	// Publishing rate control
	lastPublish time.Time
	publishRate time.Duration

	// Trigger detection
	triggerLevel float64
	lastSample   float64

	// Sample configuration
	sampleRate int

	mu sync.Mutex
}

// Global scope registry
var (
	scopeRegistry   = make(map[string]*Scope)
	scopeRegistryMu sync.RWMutex
)

// NewScope creates a new scope unit generator
func NewScope(name string, bufferSize int) *Scope {
	if bufferSize <= 0 {
		bufferSize = 2048
	}

	s := &Scope{
		id:           uuid.New().String(),
		name:         name,
		bufferSize:   bufferSize,
		buffer:       make([]float64, bufferSize),
		publishRate:  time.Second / 30, // 30Hz update rate
		triggerLevel: 0.0,
	}

	return s
}

// ID returns the scope's unique identifier
func (s *Scope) ID() string {
	return s.id
}

// Start initializes the scope and registers it
func (s *Scope) Start(ctx context.Context) error {
	scopeRegistryMu.Lock()
	scopeRegistry[s.id] = s
	scopeRegistryMu.Unlock()

	// Notify frontend about scope changes
	publishScopeList()

	return nil
}

// Stop unregisters the scope
func (s *Scope) Stop(ctx context.Context) error {
	scopeRegistryMu.Lock()
	delete(scopeRegistry, s.id)
	scopeRegistryMu.Unlock()

	// Notify frontend about scope changes
	publishScopeList()

	return nil
}

// Gen processes audio samples
func (s *Scope) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	// Get input samples from "in" edge
	in := cfg.InputSamples["in"]
	if in == nil {
		// No input, output silence
		for i := range out {
			out[i] = 0
		}
		return
	}

	// Pass through unchanged
	copy(out, in)

	// Buffer samples for visualization
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sample := range in {
		s.buffer[s.bufferIndex] = sample
		s.bufferIndex = (s.bufferIndex + 1) % s.bufferSize
	}

	// Store sample rate for publishing
	if s.sampleRate == 0 {
		s.sampleRate = cfg.SampleRateHz
	}

	// Check if it's time to publish
	now := time.Now()
	if now.Sub(s.lastPublish) >= s.publishRate {
		s.publishBuffer()
		s.lastPublish = now
	}
}

// publishBuffer sends the current buffer to the frontend
func (s *Scope) publishBuffer() {
	// Find trigger index for stable display
	triggerIndex := s.findTriggerIndex()

	// Create a copy of the buffer starting from trigger point
	samples := make([]float64, s.bufferSize)
	for i := 0; i < s.bufferSize; i++ {
		idx := (triggerIndex + i) % s.bufferSize
		samples[i] = s.buffer[idx]
	}

	// Publish scope data
	data := map[string]interface{}{
		"id":           s.id,
		"name":         s.name,
		"samples":      samples,
		"sampleRate":   s.sampleRate,
		"triggerIndex": 0, // Already adjusted in samples
		"timestamp":    time.Now().UnixMilli(),
	}

	pubsub.Publish("scope.data", data)
}

// findTriggerIndex finds a rising-edge zero crossing for stable display
func (s *Scope) findTriggerIndex() int {
	// Simple trigger: find first positive-going zero crossing
	for i := 0; i < s.bufferSize; i++ {
		idx := (s.bufferIndex + i) % s.bufferSize
		nextIdx := (idx + 1) % s.bufferSize

		current := s.buffer[idx]
		next := s.buffer[nextIdx]

		// Check for rising edge crossing trigger level
		if current <= s.triggerLevel && next > s.triggerLevel {
			return idx
		}
	}

	// No trigger found, use current position
	return s.bufferIndex
}

// publishScopeList notifies about active scopes
func publishScopeList() {
	scopeRegistryMu.RLock()
	defer scopeRegistryMu.RUnlock()

	scopes := make([]map[string]string, 0, len(scopeRegistry))
	for id, scope := range scopeRegistry {
		scopes = append(scopes, map[string]string{
			"id":   id,
			"name": scope.name,
		})
	}

	pubsub.Publish("scopes-changed", map[string]interface{}{
		"scopes": scopes,
	})
}

// Ensure Scope implements required interfaces
var _ UGen = (*Scope)(nil)
var _ Starter = (*Scope)(nil)
var _ Stopper = (*Scope)(nil)

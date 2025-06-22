# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Muscrat is a real-time computer music system and programming language for macOS. It combines a custom Lisp dialect (based on Glojure) with a desktop application built using Go, Wails v3, and React.

## Development Commands

### Build Commands
- `make` or `make app` - Build the full desktop application
- `make dev` - Run in development mode with hot-reload
- `make gen` - Generate Go interop code for Glojure bindings (required after modifying Go packages)
- `make clean` - Clean build artifacts

### Testing
- Backend: `go test ./...`
- Frontend: `cd frontend && npm test`

### Linting
- Frontend: `cd frontend && npm run lint`

### Environment Setup
The project uses Nix shell. Enter the development environment with:
```bash
nix-shell
```

## Architecture Overview

### Backend (Go)
- **Audio Engine** (`pkg/ugen/`): Real-time audio processing with unit generators
- **Language Runtime** (`pkg/stdlib/mrat/`): Glojure-based Lisp interpreter
- **Audio I/O** (`pkg/aio/`): PortAudio integration for audio input/output
- **Effects** (`pkg/effects/`): Audio effects processing (reverb, delay, filters)
- **Sample Library** (`pkg/sampler/`): Sample playback and manipulation
- **Graph Processing** (`pkg/graph/`): Audio signal flow management

### Frontend (React + Vite)
- **Code Editor**: Monaco Editor integration for live coding
- **Visualizations**: Real-time oscilloscope and spectrogram
- **Visual Synthesis**: Hydra integration for visuals
- **UI Components**: Knobs and controls for parameter adjustment

### Language (.glj files)
The language is a Lisp dialect for music synthesis. Example structure:
```clojure
(ns user
  (:require [mrat.core :refer :all]))

(play (sin 440 :mul 0.1))
```

## Important Environment Variables
- `MUSCRAT_SAMPLE_PATH`: Path to audio samples (default: `./data/samples`)
- `MUSCRAT_STDLIB_PATH`: Path to standard library (default: `./pkg/stdlib`)

## Key Directories
- `/examples/` - Example .glj files demonstrating language features
- `/data/samples/` - Built-in audio sample library
- `/pkg/stdlib/mrat/` - Core language functions and DSP primitives
- `/frontend/src/` - React application source

## Development Notes
- The project requires macOS due to audio framework dependencies
- When modifying Go packages that interface with the Lisp runtime, run `make gen` to regenerate bindings
- The development server supports hot-reload for both frontend and backend changes
- Audio processing runs in real-time, so performance is critical in DSP code
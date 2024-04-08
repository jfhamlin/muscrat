#!/bin/bash

set -e

# universal fails to build
wails build -platform darwin/arm64 -tags static -ldflags "-s -w"
cp -r data/samples build/bin/muscrat.app/Contents/Resources/samples

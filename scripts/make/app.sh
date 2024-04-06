#!/bin/bash

set -e

# universal fails to build
wails build -platform darwin/arm64
cp -r data/samples build/bin/muscrat.app/Contents/Resources/samples

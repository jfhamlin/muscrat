#!/bin/bash

set -e

wails build -platform darwin/arm64
cp -r data/samples build/bin/muscrat.app/Contents/Resources/samples

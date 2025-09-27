
# environment variables

MUSCRAT_SAMPLE_PATH ?= ./data/samples
MUSCRAT_STDLIB_PATH ?= ./pkg/stdlib

.PHONY: dev
dev: wails
	@MUSCRAT_SAMPLE_PATH=$(MUSCRAT_SAMPLE_PATH) \
	 MUSCRAT_STDLIB_PATH=$(MUSCRAT_STDLIB_PATH) \
	 wails3 dev

.PHONY: app
app: wails
	@./scripts/make/app.sh

.PHONY: gen
gen:
	@GOARCH=$(shell go env GOARCH) go run github.com/glojurelang/glojure/cmd/gen-import-interop -packages=github.com/jfhamlin/muscrat/pkg/ugen,github.com/jfhamlin/muscrat/pkg/wavtabs,github.com/jfhamlin/muscrat/pkg/osc,github.com/jfhamlin/muscrat/pkg/stochastic,github.com/jfhamlin/muscrat/pkg/effects,github.com/jfhamlin/muscrat/pkg/mod,github.com/jfhamlin/muscrat/pkg/sampler,github.com/jfhamlin/muscrat/pkg/aio,github.com/jfhamlin/muscrat/pkg/graph,github.com/jfhamlin/muscrat/pkg/pattern,github.com/jfhamlin/freeverb-go,github.com/jfhamlin/muscrat/pkg/slice,github.com/jfhamlin/muscrat/pkg/conf > pkg/gen/gljimports/gljimports.go

.PHONY: wails
wails:
	@which wails3 2>&1 > /dev/null || go install github.com/wailsapp/wails/v3/cmd/wails3@latest

.PHONY: clean
clean:
	@rm -rf build/bin

.PHONY: typecheck
typecheck:
	@cd frontend && npx tsc --noEmit

# run-race: # disable checkptr because glojure uses an unsafe method to support dynamic vars.
# 	@go run -gcflags=all=-d=checkptr=0 -race cmd/mrat/main.go  $(SCRIPT)

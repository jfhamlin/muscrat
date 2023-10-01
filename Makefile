
.PHONY: gen
gen:
	@go run github.com/glojurelang/glojure/cmd/gen-import-interop -packages=github.com/jfhamlin/muscrat/pkg/ugen,github.com/jfhamlin/muscrat/pkg/wavtabs,github.com/jfhamlin/muscrat/pkg/osc,github.com/jfhamlin/muscrat/pkg/stochastic,github.com/jfhamlin/muscrat/pkg/effects,github.com/jfhamlin/muscrat/pkg/mod,github.com/jfhamlin/muscrat/pkg/sampler,github.com/jfhamlin/muscrat/pkg/aio,github.com/jfhamlin/muscrat/pkg/graph,github.com/jfhamlin/muscrat/pkg/pattern,github.com/jfhamlin/freeverb-go,github.com/jfhamlin/muscrat/pkg/slice > pkg/gen/gljimports/gljimports.go

run:
	@go run cmd/mrat/main.go $(SCRIPT)

run-race: # disable checkptr because glojure uses an unsafe method to support dynamic vars.
	@go run -gcflags=all=-d=checkptr=0 -race cmd/mrat/main.go  $(SCRIPT)

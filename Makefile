
.PHONY: gen
gen:
	@go run ../glojure/cmd/gen-import-interop/main.go -packages=github.com/jfhamlin/muscrat/pkg/ugen,github.com/jfhamlin/muscrat/pkg/wavtabs,github.com/jfhamlin/muscrat/pkg/graph > pkg/gen/gljimports/gljimports.go

// Package stdlib provides the standard library for the mratlang language.
package stdlib

import (
	"embed"
)

//go:embed mrat
var StdLib embed.FS

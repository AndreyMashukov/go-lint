// Package nolintdirective forbids linter suppression directives.
package nolintdirective

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nolintdirective",
	Doc:  "forbids //nolint, //lint:ignore, //staticcheck:ignore, //revive:disable, //go:linkname directives",
	Run:  run,
}

var forbiddenPrefixes = []string{
	"//nolint",
	"// nolint",
	"//lint:ignore",
	"// lint:ignore",
	"//staticcheck:ignore",
	"// staticcheck:ignore",
	"//revive:disable",
	"// revive:disable",
	"//go:linkname",
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				text := c.Text
				for _, p := range forbiddenPrefixes {
					if strings.HasPrefix(text, p) {
						pass.Reportf(c.Slash, "linter suppression is forbidden — fix the underlying issue")
						break
					}
				}
			}
		}
	}
	return nil, nil
}

// Package notodo forbids TODO/FIXME/XXX/HACK markers outright.
package notodo

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "notodo",
	Doc:  "forbids TODO/FIXME/XXX/HACK markers outright — implement it now or track it in an issue, do not leave a stub",
	Run:  run,
}

var markers = []string{"TODO", "FIXME", "XXX", "HACK"}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if m := openingMarker(c.Text); m != "" {
					pass.Reportf(c.Slash, "%s marker is forbidden — implement it now, do not leave a stub", m)
				}
			}
		}
	}
	return nil, nil
}

// openingMarker returns the marker a comment opens with at a word boundary, or
// "" when the comment merely mentions a marker mid-sentence (such as this
// package's own docs).
func openingMarker(text string) string {
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	text = strings.TrimLeft(text, "/!* \t")
	text = strings.TrimSpace(text)
	for _, m := range markers {
		if rest, ok := strings.CutPrefix(text, m); ok {
			if rest == "" || !isAlnum(rest[0]) {
				return m
			}
		}
	}
	return ""
}

func isAlnum(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

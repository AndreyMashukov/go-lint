// Package nodbmutintest forbids direct DB mutation in tests.
package nodbmutintest

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nodbmutintest",
	Doc:  "forbids direct DB mutation (INSERT/UPDATE/DELETE/TRUNCATE/DROP/ALTER) in tests — drive state through production code paths",
	Run:  run,
}

var execMethods = map[string]bool{
	"Exec":         true,
	"ExecContext":  true,
	"Query":        true,
	"QueryContext": true,
	"QueryRow":     true,
}

var mutatingKeywords = []string{
	"INSERT ",
	"UPDATE ",
	"DELETE ",
	"TRUNCATE ",
	"DROP ",
	"ALTER ",
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, "_test.go") && !strings.Contains(filename, "/tests/") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !execMethods[sel.Sel.Name] {
				return true
			}
			for _, arg := range call.Args {
				lit, ok := arg.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				upper := strings.ToUpper(strings.Trim(lit.Value, "`\""))
				for _, kw := range mutatingKeywords {
					if strings.Contains(upper, kw) {
						pass.Reportf(call.Pos(), "direct DB mutation in test — drive state through production code paths")
						return true
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

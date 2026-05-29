// Package notypeonlyassert forbids type-only or existence-only assertions in tests.
package notypeonlyassert

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "notypeonlyassert",
	Doc:  "forbids type-only and existence-only assertions like NotNil, IsType, NotEmpty — assert exact value",
	Run:  run,
}

var forbiddenAssertions = map[string]bool{
	"IsType":     true,
	"NotNil":     true,
	"NotEmpty":   true,
	"NotZero":    true,
	"Implements": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, "_test.go") {
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
			if !forbiddenAssertions[sel.Sel.Name] {
				return true
			}
			recvName := receiverPackageName(sel.X)
			if recvName != "assert" && recvName != "require" {
				return true
			}
			pass.Reportf(call.Pos(), "type-only or existence-only assertion (%s.%s) — assert exact value", recvName, sel.Sel.Name)
			return true
		})
	}
	return nil, nil
}

func receiverPackageName(x ast.Expr) string {
	ident, ok := x.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

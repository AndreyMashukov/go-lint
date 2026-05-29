// Package notimenow forbids direct time.Now/Since/Until outside clock package.
package notimenow

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "notimenow",
	Doc:  "forbids direct time.Now/time.Since/time.Until outside clock package — inject Clock interface for testability",
	Run:  run,
}

var forbiddenFuncs = map[string]bool{
	"Now":   true,
	"Since": true,
	"Until": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg != nil && pass.Pkg.Name() == "clock" {
		return nil, nil
	}
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		base := filepath.Base(filename)
		if base == "clock.go" || strings.HasSuffix(base, "_test.go") {
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
			if !forbiddenFuncs[sel.Sel.Name] {
				return true
			}
			if !isTimePackage(pass, sel.X) {
				return true
			}
			pass.Reportf(call.Pos(), "direct time.%s — inject Clock interface for testability", sel.Sel.Name)
			return true
		})
	}
	return nil, nil
}

func isTimePackage(pass *analysis.Pass, x ast.Expr) bool {
	ident, ok := x.(*ast.Ident)
	if !ok {
		return false
	}
	if pass.TypesInfo == nil {
		return ident.Name == "time"
	}
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return ident.Name == "time"
	}
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "time"
}

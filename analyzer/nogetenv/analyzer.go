// Package nogetenv forbids os.Getenv outside config packages.
package nogetenv

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nogetenv",
	Doc:  "forbids os.Getenv and os.LookupEnv outside config packages — inject config struct instead",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if isConfigPackage(pass) {
		return nil, nil
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name != "Getenv" && sel.Sel.Name != "LookupEnv" {
				return true
			}
			if !isOsPackage(pass, sel.X) {
				return true
			}
			pass.Reportf(call.Pos(), "os.%s outside config package — inject config struct instead", sel.Sel.Name)
			return true
		})
	}
	return nil, nil
}

func isConfigPackage(pass *analysis.Pass) bool {
	if pass.Pkg == nil {
		return false
	}
	path := pass.Pkg.Path()
	name := pass.Pkg.Name()
	if name == "config" {
		return true
	}
	return strings.Contains(path, "/config/") || strings.HasSuffix(path, "/config")
}

func isOsPackage(pass *analysis.Pass, x ast.Expr) bool {
	ident, ok := x.(*ast.Ident)
	if !ok {
		return false
	}
	if pass.TypesInfo == nil {
		return ident.Name == "os"
	}
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return ident.Name == "os"
	}
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "os"
}

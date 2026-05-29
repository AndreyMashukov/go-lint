// Package noerrorwrapbanality forbids fmt.Errorf wrappers that add no context.
package noerrorwrapbanality

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noerrorwrapbanality",
	Doc:  "forbids fmt.Errorf wrappers without added context — return err directly or add real context",
	Run:  run,
}

var banalRe = regexp.MustCompile(`(?i)^(failed to|error|cannot|could not|unable to)\s+\w+:?\s*%[ws]$`)

func run(pass *analysis.Pass) (interface{}, error) {
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
			if sel.Sel.Name != "Errorf" {
				return true
			}
			if !isFmtPackage(pass, sel.X) {
				return true
			}
			if len(call.Args) < 1 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			format := strings.Trim(lit.Value, "`\"")
			if len(call.Args) != 2 {
				return true
			}
			if !banalRe.MatchString(strings.TrimSpace(format)) {
				return true
			}
			pass.Reportf(call.Pos(), "fmt.Errorf wrapper without added context — return err directly or add real context")
			return true
		})
	}
	return nil, nil
}

func isFmtPackage(pass *analysis.Pass, x ast.Expr) bool {
	ident, ok := x.(*ast.Ident)
	if !ok {
		return false
	}
	if pass.TypesInfo == nil {
		return ident.Name == "fmt"
	}
	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return ident.Name == "fmt"
	}
	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "fmt"
}

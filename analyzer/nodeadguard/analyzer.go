// Package nodeadguard forbids nil-checks on values that cannot be nil.
package nodeadguard

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name:             "nodeadguard",
	Doc:              "forbids nil-check on value type that cannot be nil — dead defensive guard",
	Run:              run,
	RunDespiteErrors: true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.TypesInfo == nil {
		return nil, nil
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}
			bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
			if !ok {
				return true
			}
			if bin.Op != token.EQL && bin.Op != token.NEQ {
				return true
			}
			var target ast.Expr
			if isNilIdent(bin.Y) {
				target = bin.X
			} else if isNilIdent(bin.X) {
				target = bin.Y
			} else {
				return true
			}
			tv, ok := pass.TypesInfo.Types[target]
			if !ok {
				return true
			}
			t := tv.Type
			if t == nil {
				return true
			}
			if canBeNil(t) {
				return true
			}
			pass.Reportf(bin.Pos(), "nil-check on value type that cannot be nil — dead defensive guard")
			return true
		})
	}
	return nil, nil
}

func isNilIdent(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func canBeNil(t types.Type) bool {
	switch u := t.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		_ = u
		return true
	}
	return false
}

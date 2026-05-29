// Package noredundantif forbids redundant if-return patterns.
package noredundantif

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noredundantif",
	Doc:  "forbids redundant if-return — use 'return <cond>' (or its negation)",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}
			for i := 0; i+1 < len(block.List); i++ {
				ifStmt, ok := block.List[i].(*ast.IfStmt)
				if !ok || ifStmt.Else != nil {
					continue
				}
				retIf, ok := singleReturn(ifStmt.Body)
				if !ok {
					continue
				}
				nextRet, ok := block.List[i+1].(*ast.ReturnStmt)
				if !ok {
					continue
				}
				if len(retIf.Results) != 1 || len(nextRet.Results) != 1 {
					continue
				}
				b1, ok1 := boolValue(retIf.Results[0])
				b2, ok2 := boolValue(nextRet.Results[0])
				if !ok1 || !ok2 {
					continue
				}
				if b1 == b2 {
					continue
				}
				pass.Reportf(ifStmt.Pos(), "redundant if-return — use 'return <cond>' (or its negation)")
			}
			return true
		})
	}
	return nil, nil
}

func singleReturn(body *ast.BlockStmt) (*ast.ReturnStmt, bool) {
	if body == nil || len(body.List) != 1 {
		return nil, false
	}
	ret, ok := body.List[0].(*ast.ReturnStmt)
	return ret, ok
}

func boolValue(e ast.Expr) (bool, bool) {
	ident, ok := e.(*ast.Ident)
	if !ok {
		return false, false
	}
	switch ident.Name {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}

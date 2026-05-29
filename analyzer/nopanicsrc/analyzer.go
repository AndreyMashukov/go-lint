// Package nopanicsrc forbids panic() in production code.
package nopanicsrc

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nopanicsrc",
	Doc:  "forbids panic() in production code — return error instead",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		isMainPkg := file.Name.Name == "main"
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if fn.Name.Name == "init" {
				continue
			}
			if isMainPkg && fn.Name.Name == "main" {
				continue
			}
			deferRanges := collectDeferRanges(fn.Body)
			checkBody(pass, fn.Body, deferRanges)
		}
	}
	return nil, nil
}

func collectDeferRanges(body *ast.BlockStmt) [][2]token.Pos {
	var ranges [][2]token.Pos
	ast.Inspect(body, func(n ast.Node) bool {
		d, ok := n.(*ast.DeferStmt)
		if !ok {
			return true
		}
		fl, ok := d.Call.Fun.(*ast.FuncLit)
		if !ok || fl.Body == nil {
			return true
		}
		ranges = append(ranges, [2]token.Pos{fl.Body.Lbrace, fl.Body.Rbrace})
		return true
	})
	return ranges
}

func checkBody(pass *analysis.Pass, body *ast.BlockStmt, deferRanges [][2]token.Pos) {
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "panic" {
			return true
		}
		if isInDefer(call.Pos(), deferRanges) {
			return true
		}
		pass.Reportf(call.Pos(), "panic in production code — return error instead")
		return true
	})
}

func isInDefer(pos token.Pos, ranges [][2]token.Pos) bool {
	for _, r := range ranges {
		if pos > r[0] && pos < r[1] {
			return true
		}
	}
	return false
}

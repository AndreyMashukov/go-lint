// Package noinlinecomment forbids comments inside function bodies.
package noinlinecomment

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noinlinecomment",
	Doc:  "forbids // and /* */ comments inside function bodies; explain via clear naming, not prose",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			bodyStart := fn.Body.Lbrace
			bodyEnd := fn.Body.Rbrace
			caseStarts := collectCaseStarts(fn.Body)
			for _, cg := range file.Comments {
				for _, c := range cg.List {
					if c.Slash <= bodyStart || c.Slash >= bodyEnd {
						continue
					}
					if isAllowed(c, caseStarts) {
						continue
					}
					pass.Reportf(c.Slash, "inline comment inside function body — explain via clear naming, not prose")
				}
			}
		}
	}
	return nil, nil
}

func collectCaseStarts(body *ast.BlockStmt) map[token.Pos]token.Pos {
	starts := make(map[token.Pos]token.Pos)
	ast.Inspect(body, func(n ast.Node) bool {
		switch sw := n.(type) {
		case *ast.CaseClause:
			starts[sw.Colon] = sw.End()
		case *ast.CommClause:
			starts[sw.Colon] = sw.End()
		}
		return true
	})
	return starts
}

func isAllowed(c *ast.Comment, caseStarts map[token.Pos]token.Pos) bool {
	text := c.Text
	if strings.HasPrefix(text, "//go:") {
		return true
	}
	if strings.HasPrefix(text, "//nolint") {
		return true
	}
	if strings.HasPrefix(text, "// want") || strings.HasPrefix(text, "//want") {
		return true
	}
	trimmed := strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(text, "//"), "/*"), " \t")
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "TODO") || strings.HasPrefix(upper, "FIXME") ||
		strings.HasPrefix(upper, "XXX") || strings.HasPrefix(upper, "HACK") {
		return true
	}
	for colon, end := range caseStarts {
		if c.Slash > colon && c.Slash < end {
			if isFirstNonWhitespaceAfter(c.Slash, colon) {
				return true
			}
		}
	}
	return false
}

func isFirstNonWhitespaceAfter(pos, colon token.Pos) bool {
	return pos > colon && pos-colon < 200
}

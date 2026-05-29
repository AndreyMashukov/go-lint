// Package noenvbranch forbids runtime branching on environment strings.
package noenvbranch

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noenvbranch",
	Doc:  "forbids runtime branching on env strings like 'prod', 'dev', 'test' — production code must behave identically in all environments",
	Run:  run,
}

var envLiterals = map[string]bool{
	"prod":        true,
	"production":  true,
	"dev":         true,
	"development": true,
	"test":        true,
	"testing":     true,
	"stage":       true,
	"staging":     true,
	"local":       true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			bin, ok := n.(*ast.BinaryExpr)
			if !ok {
				return true
			}
			if bin.Op != token.EQL && bin.Op != token.NEQ {
				return true
			}
			if isEnvLiteral(bin.X) || isEnvLiteral(bin.Y) {
				pass.Reportf(bin.Pos(), "runtime environment branching — production code must behave identically in all environments")
			}
			return true
		})
	}
	return nil, nil
}

func isEnvLiteral(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	v := strings.Trim(lit.Value, "`\"")
	return envLiterals[strings.ToLower(v)]
}

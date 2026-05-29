// Package nosilentfallback forbids silent defaults for missing values:
// cmp.Or(x, literal) and if-block fallbacks like `if x == "" { x = "default" }`.
package nosilentfallback

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "nosilentfallback",
	Doc: "forbids silent default patterns: cmp.Or(x, <literal>) and " +
		"`if x == \"\" { x = ... }` / `if x == 0 { x = ... }` / " +
		"`if x == nil { x = ... }` post-read fallbacks. Sibling rules: " +
		"`no-silent-fallback` (eslint-plugin-mess-detector), " +
		"`NoSilentFallbackRector` (rector-php-rules), " +
		"`no_silent_fallback` (rust-lint).",
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if isTestFile(pass.Fset.Position(file.Pos()).Filename) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CallExpr:
				if pos := cmpOrWithLiteral(v); pos != token.NoPos {
					pass.Reportf(pos,
						"cmp.Or(...) with a literal argument is a silent fallback — "+
							"validate the input at its source or let it crash")
				}
			case *ast.IfStmt:
				if pos := ifFallback(v); pos != token.NoPos {
					pass.Reportf(pos,
						"silent fallback inside `if` — assign explicitly after an "+
							"existence check or let it crash")
				}
			}
			return true
		})
	}
	return nil, nil
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// cmpOrWithLiteral returns the position to report when call is `cmp.Or(...)`
// with at least one literal argument; token.NoPos otherwise.
func cmpOrWithLiteral(call *ast.CallExpr) token.Pos {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return token.NoPos
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return token.NoPos
	}
	if pkg.Name != "cmp" || sel.Sel.Name != "Or" {
		return token.NoPos
	}
	for _, arg := range call.Args {
		if isLiteralArg(arg) {
			return call.Pos()
		}
	}
	return token.NoPos
}

// isLiteralArg returns true if arg is a literal-shaped expression that signals
// a fallback default: BasicLit, nil/true/false ident, empty composite literal,
// string-concat of literals, or a parenthesised wrapper of any of those.
func isLiteralArg(arg ast.Expr) bool {
	switch v := arg.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		switch v.Name {
		case "nil", "true", "false":
			return true
		}
		return false
	case *ast.CompositeLit:
		return len(v.Elts) == 0
	case *ast.ParenExpr:
		return isLiteralArg(v.X)
	case *ast.UnaryExpr:
		if v.Op == token.SUB || v.Op == token.ADD {
			return isLiteralArg(v.X)
		}
	}
	return false
}

// ifFallback returns the position to report when stmt is a one-statement
// `if ident == <zero-literal> { <same-ident> = <expr> }` body, where
// zero-literal is "", 0, nil, or false. Returns NoPos otherwise.
func ifFallback(stmt *ast.IfStmt) token.Pos {
	if stmt.Init != nil || stmt.Else != nil {
		return token.NoPos
	}
	bin, ok := stmt.Cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.EQL {
		return token.NoPos
	}
	condIdent, ok := zeroComparison(bin)
	if !ok {
		return token.NoPos
	}
	if len(stmt.Body.List) != 1 {
		return token.NoPos
	}
	assign, ok := stmt.Body.List[0].(*ast.AssignStmt)
	if !ok {
		return token.NoPos
	}
	if assign.Tok != token.ASSIGN {
		return token.NoPos
	}
	if len(assign.Lhs) != 1 {
		return token.NoPos
	}
	lhs, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return token.NoPos
	}
	if lhs.Name != condIdent {
		return token.NoPos
	}
	return stmt.If
}

// zeroComparison returns the identifier name when bin compares an identifier
// against a zero literal ("", 0, nil, false) in either order, ok=false otherwise.
func zeroComparison(bin *ast.BinaryExpr) (string, bool) {
	if ident, ok := bin.X.(*ast.Ident); ok && isZeroLiteral(bin.Y) {
		return ident.Name, true
	}
	if ident, ok := bin.Y.(*ast.Ident); ok && isZeroLiteral(bin.X) {
		return ident.Name, true
	}
	return "", false
}

func isZeroLiteral(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		switch v.Kind {
		case token.STRING:
			return v.Value == `""` || v.Value == "``"
		case token.INT, token.FLOAT:
			return v.Value == "0" || v.Value == "0.0"
		}
	case *ast.Ident:
		return v.Name == "nil" || v.Name == "false"
	}
	return false
}

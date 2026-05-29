// Package norobotgodoc forbids godoc that tautologically restates the function signature.
package norobotgodoc

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "norobotgodoc",
	Doc:  "forbids godoc that tautologically restates the function signature — describe behavior or omit",
	Run:  run,
}

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "of": true, "to": true,
	"for": true, "with": true, "from": true, "and": true, "or": true,
	"is": true, "returns": true, "return": true, "gets": true, "get": true,
	"sets": true, "set": true, "this": true, "function": true, "it": true,
	"new": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if !fn.Name.IsExported() {
				continue
			}
			if fn.Doc == nil {
				continue
			}
			if isTautological(fn.Name.Name, fn.Doc.Text()) {
				pass.Reportf(fn.Pos(), "godoc tautologically restates function signature — describe behavior or omit")
			}
		}
	}
	return nil, nil
}

func isTautological(name, docText string) bool {
	docText = strings.TrimSpace(docText)
	if docText == "" {
		return false
	}
	words := splitWords(docText)
	if len(words) == 0 {
		return false
	}
	if !strings.EqualFold(words[0], name) {
		return false
	}
	rest := words[1:]
	nameTokens := splitCamel(name)
	var meaningful []string
	for _, w := range rest {
		lw := strings.ToLower(w)
		if stopWords[lw] {
			continue
		}
		if containsToken(nameTokens, lw) {
			continue
		}
		if isVerbForm(nameTokens, lw) {
			continue
		}
		meaningful = append(meaningful, lw)
	}
	return len(meaningful) <= 2
}

func splitWords(s string) []string {
	var words []string
	for _, line := range strings.Split(s, "\n") {
		for _, tok := range strings.FieldsFunc(line, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			words = append(words, tok)
		}
	}
	return words
}

func splitCamel(name string) []string {
	var parts []string
	var current []rune
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) && len(current) > 0 {
			parts = append(parts, strings.ToLower(string(current)))
			current = nil
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		parts = append(parts, strings.ToLower(string(current)))
	}
	return parts
}

func containsToken(tokens []string, w string) bool {
	for _, t := range tokens {
		if t == w {
			return true
		}
	}
	return false
}

func isVerbForm(tokens []string, w string) bool {
	for _, t := range tokens {
		if t+"s" == w || t+"es" == w || t+"ed" == w || t+"d" == w || t+"ing" == w {
			return true
		}
		if len(t) > 1 && t[len(t)-1] == 'e' && t[:len(t)-1]+"ing" == w {
			return true
		}
		if len(t) > 1 && t[len(t)-1] == 'y' && t[:len(t)-1]+"ies" == w {
			return true
		}
	}
	return false
}

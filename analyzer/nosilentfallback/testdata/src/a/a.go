package a

import "cmp"

func CmpOrLiteralString(name string) string {
	return cmp.Or(name, "default") // want `cmp.Or\(\.\.\.\) with a literal argument`
}

func CmpOrLiteralInt(n int) int {
	return cmp.Or(n, 1) // want `cmp.Or\(\.\.\.\) with a literal argument`
}

func CmpOrNil(p *int) *int {
	return cmp.Or(p, nil) // want `cmp.Or\(\.\.\.\) with a literal argument`
}

func CmpOrChainedSort(byName, byID, byCreated int) int {
	// All non-literal — multi-field sort comparator, legitimate.
	return cmp.Or(byName, byID, byCreated)
}

func IfStringFallback(s string) string {
	if s == "" { // want `silent fallback inside .if.`
		s = "default"
	}
	return s
}

func IfIntFallback(n int) int {
	if n == 0 { // want `silent fallback inside .if.`
		n = 1
	}
	return n
}

func IfNilFallback(p *int, fallback *int) *int {
	if p == nil { // want `silent fallback inside .if.`
		p = fallback
	}
	return p
}

func IfBoolFallback(flag bool) bool {
	if flag == false { // want `silent fallback inside .if.`
		flag = true
	}
	return flag
}

func IfEmptyCheckExplicit(s string) error {
	if s == "" {
		return ErrMissing
	}
	return nil
}

func IfNilCheckExplicit(p *int) error {
	if p == nil {
		panic("p is required")
	}
	return nil
}

func IfEmptyCheckReturn(s string, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func IfWithElseBranch(n int) int {
	if n == 0 {
		n = 1
	} else {
		n = n + 1
	}
	return n
}

func IfWithInit(m map[string]string) string {
	if v, ok := m["k"]; ok {
		return v
	}
	return ""
}

var ErrMissing = errMissing()

func errMissing() error { return nil }

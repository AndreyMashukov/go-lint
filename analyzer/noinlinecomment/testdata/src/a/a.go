package a

// PackageLevel is fine.
func Bad(x int) int {
	// first we double // want `inline comment inside function body`
	y := x * 2
	/* then we add one */ // want `inline comment inside function body`
	return y + 1
}

func OK(x int) int {
	return x*2 + 1
}

func WithDirective() {
	//go:noinline
	_ = 1
}

func WithTodo() {
	// TODO(@me): something
	_ = 1
}

func WithSwitchCase(x int) string {
	switch x {
	case 1:
		// initialize default
		return "one"
	case 2:
		return "two"
	}
	return ""
}

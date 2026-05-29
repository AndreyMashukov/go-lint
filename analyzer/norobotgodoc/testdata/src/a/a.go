package a

// Add adds two numbers.
func Add(a, b int) int { return a + b } // want `godoc tautologically restates function signature`

// GetName returns the name.
func GetName() string { return "" } // want `godoc tautologically restates function signature`

// Process processes the input applying domain-specific rules with retry on transient errors.
func Process(input string) string { return input }

// Multiply multiplies x by y and returns the float result rounded to nearest tick.
func Multiply(x, y float64) float64 { return x * y }

func Unexported() {}

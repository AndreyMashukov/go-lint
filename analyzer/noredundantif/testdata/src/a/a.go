package a

func Bad1(x int) bool {
	if x > 0 { // want `redundant if-return`
		return true
	}
	return false
}

func Bad2(x int) bool {
	if x == 0 { // want `redundant if-return`
		return false
	}
	return true
}

func OK1(x int) bool {
	return x > 0
}

func OK2(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}

func OK3(x int) bool {
	if x > 0 {
		return true
	}
	return true
}

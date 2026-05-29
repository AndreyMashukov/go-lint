package a

type plainStruct struct {
	X int
	Y int
}

func Bad1(x int) bool {
	if x == nil { // want `nil-check on value type that cannot be nil`
		return true
	}
	return false
}

func Bad2(s string) bool {
	if s == nil { // want `nil-check on value type that cannot be nil`
		return true
	}
	return false
}

func Bad3(p plainStruct) bool {
	if p == nil { // want `nil-check on value type that cannot be nil`
		return true
	}
	return false
}

func OKPointer(p *plainStruct) bool {
	return p == nil
}

func OKSlice(s []int) bool {
	return s == nil
}

func OKMap(m map[string]int) bool {
	return m == nil
}

func OKInterface(i interface{}) bool {
	return i == nil
}

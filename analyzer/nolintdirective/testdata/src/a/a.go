package a

//nolint:errcheck // want `linter suppression is forbidden`
func Bad1() {}

//lint:ignore U1000 unused // want `linter suppression is forbidden`
func Bad2() {}

//revive:disable // want `linter suppression is forbidden`
func Bad3() {}

func OK() {}

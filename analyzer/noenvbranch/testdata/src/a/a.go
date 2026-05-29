package a

func Bad(env string) bool {
	return env == "prod" // want `runtime environment branching`
}

func Bad2(env string) bool {
	return env != "development" // want `runtime environment branching`
}

func Bad3(env string) bool {
	return "staging" == env // want `runtime environment branching`
}

func OK(env string) bool {
	return env == "frobnicate"
}

package a

import "os"

func Bad() string {
	return os.Getenv("HOME") // want `os.Getenv outside config package`
}

func Bad2() (string, bool) {
	return os.LookupEnv("PATH") // want `os.LookupEnv outside config package`
}

func OK() string {
	return "static"
}

package a

import (
	"errors"
	"fmt"
)

var sentinel = errors.New("x")

func Bad1() error {
	return fmt.Errorf("failed to read: %w", sentinel) // want `fmt.Errorf wrapper without added context`
}

func Bad2() error {
	return fmt.Errorf("cannot parse: %w", sentinel) // want `fmt.Errorf wrapper without added context`
}

func Bad3() error {
	return fmt.Errorf("unable to fetch %w", sentinel) // want `fmt.Errorf wrapper without added context`
}

func OK1() error {
	return fmt.Errorf("read /etc/foo for user %d: %w", 7, sentinel)
}

func OK2() error {
	return fmt.Errorf("config file is %s: %w", "missing", sentinel)
}

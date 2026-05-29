package a

func Bad() {
	panic("nope") // want `panic in production code`
}

func init() {
	panic("startup failed")
}

func WithDeferRecover() {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
}

func OK() error {
	return nil
}

package a

import "time"

func Bad() time.Time {
	return time.Now() // want `direct time.Now`
}

func Bad2(t time.Time) time.Duration {
	return time.Since(t) // want `direct time.Since`
}

func Bad3(t time.Time) time.Duration {
	return time.Until(t) // want `direct time.Until`
}

func OK(t time.Time) time.Time {
	return t.Add(time.Second)
}

package timing

import (
	"runtime"
	"time"
)

// callerName returns the name of the function skip frames up the call stack.
func callerName(skip int) string {
	const unknown = "unknown"
	pcs := make([]uintptr, 1)
	n := runtime.Callers(skip+2, pcs)
	if n < 1 {
		return unknown
	}
	frame, _ := runtime.CallersFrames(pcs).Next()
	if frame.Function == "" {
		return unknown
	}
	return frame.Function
}

func Block() func(cb func(string, time.Duration)) {
	name := callerName(1)
	start := time.Now()
	return func(cb func(string, time.Duration)) {
		cb(name, time.Since(start))
	}
}

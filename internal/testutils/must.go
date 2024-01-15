package testutils

func Must[I, R any](fn func(I) (R, error)) func(I) R {
	return func(value I) R {
		result, err := fn(value)
		if err != nil {
			panic(err)
		}

		return result
	}
}

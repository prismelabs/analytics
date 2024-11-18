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

func MustNoArg[R any](fn func() (R, error)) func() R {
	return func() R {
		result, err := fn()
		if err != nil {
			panic(err)
		}

		return result
	}
}

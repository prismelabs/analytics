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

func Must2[I1, I2, R any](fn func(I1, I2) (R, error)) func(I1, I2) R {
	return func(arg1 I1, arg2 I2) R {
		result, err := fn(arg1, arg2)
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

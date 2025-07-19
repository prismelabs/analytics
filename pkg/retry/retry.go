package retry

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// NeverCancel always return false.
func NeverCancel(_ error) bool { return false }

// CancelOnContextError returns true if provided error is a context.Canceled error
// or context.DeadlineExceeded error.
func CancelOnContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Retry runs the given function with a non deterministic linear backoff until
// it succeed or cancel returns true.
func LinearBackoff(maxRetry uint, base time.Duration, fn func(uint) error, cancel func(error) bool) error {
	var fnErrors []error

	for retry := range maxRetry {
		err := fn(retry)
		if err != nil {

			fnErrors = append(fnErrors, err)
			if cancel(err) {
				break
			}
			time.Sleep(time.Duration(retry+1) * base)
			continue
		}

		return nil
	}

	return errors.Join(fnErrors...)
}

// Retry runs the given function with a non deterministic linear backoff until
// it succeed or cancel returns true.
func LinearRandomBackoff(maxRetry uint, base time.Duration, fn func(uint) error, cancel func(error) bool) error {
	var fnErrors []error

	for retry := range maxRetry {
		err := fn(retry)
		if err != nil {

			fnErrors = append(fnErrors, err)
			if cancel(err) {
				break
			}
			time.Sleep((time.Duration(rand.Intn(3)) + time.Duration(retry+1)) * base)
			continue
		}

		return nil
	}

	return errors.Join(fnErrors...)
}

// Retry runs the given function with a non deterministic exponential backoff
// until it succeed or cancel returns true.
func ExponentialRandomBackoff(maxRetry uint, base time.Duration, fn func(uint) error, cancel func(error) bool) error {
	var fnErrors []error

	if uint(1<<maxRetry) == 0 {
		panic("invalid max retry argument")
	}

	for retry := range maxRetry {
		err := fn(retry)
		if err != nil {

			fnErrors = append(fnErrors, err)
			if cancel(err) {
				break
			}
			time.Sleep((time.Duration(rand.Intn(3)) + time.Duration(1<<retry)) * base)
			continue
		}

		return nil
	}

	return errors.Join(fnErrors...)
}

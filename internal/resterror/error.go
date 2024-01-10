package resterror

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// RestError is JSON serializable error type that can be used with any web framework.
type RestError struct {
	Code string
}

// RestError implements the error interface.
func (re RestError) Error() string {
	return fmt.Sprintf("RestError(%v)", re.Code)
}

// Create a new error joining a RestError with the given code and errs.
func New(code string, errs ...error) error {
	errs = append(errs, RestError{code})
	return errors.Join(errs...)
}

// NewFiber returns a new error with the given error code, response status and
// error description.
func NewFiber(code string, status int, desc string) error {
	return New(code, fiber.NewError(status, desc))
}

// NewFiberf works the same as NewFiber but adds formatting support.
func NewFiberf(code string, status int, format string, args ...any) error {
	return NewFiber(code, status, fmt.Sprintf(format, args...))
}

const InternalServerErrorCode = "InternalServerError"

// FiberErrorHandler is a fiber error handler that returns JSON encoded response.
// It supports RestError and *fiber.Error and errors.Join of them. If error handled
// contains none of them, a generic InternalServerError is sent without disclosing
// it. Fiber error with a status >= 500 are also hidden.
func FiberErrorHandler(c *fiber.Ctx, err error) error {
	responseBody := struct {
		Code string `json:"error_code"`
		Desc string `json:"error_description"`
	}{
		Code: InternalServerErrorCode,
		Desc: "internal server error, check server logs for more information",
	}
	c.Response().SetStatusCode(fiber.StatusInternalServerError)

	var restErr RestError
	if errors.As(err, &restErr) {
		c.Response().SetStatusCode(fiber.StatusBadRequest)
		responseBody.Code = restErr.Code
		responseBody.Desc = ""
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		c.Response().SetStatusCode(fiberErr.Code)
		if fiberErr.Code < 500 {
			responseBody.Desc = fiberErr.Message
		}

		if responseBody.Code == InternalServerErrorCode {
			responseBody.Code = "UnexpectedError"
		}
	}

	return c.JSON(responseBody)
}

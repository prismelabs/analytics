package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/resterror"
)

func RestError(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		// Handler error.
		handlerErr := resterror.FiberErrorHandler(c, err)
		// If failed to handle error, return it.
		if handlerErr != nil {
			return handlerErr
		}

		// Return error so it can be logged.
		return err
	}

	return err
}

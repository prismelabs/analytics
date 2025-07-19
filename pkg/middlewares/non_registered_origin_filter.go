package middlewares

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
)

// NonRegisteredOriginFilter returns a middleware that filter request with non
// registered origins.
func NonRegisteredOriginFilter(originRegistry originregistry.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := utils.UnsafeString(c.Request().Header.Peek(fiber.HeaderOrigin))
		origin, found := strings.CutPrefix(origin, "https://")
		if !found {
			origin = strings.TrimPrefix(origin, "http://")
		}

		portIndex := strings.LastIndexByte(origin, ':')
		if portIndex > 0 {
			origin = origin[:portIndex]
		}

		registered, err := originRegistry.IsOriginRegistered(c.UserContext(), origin)
		if err != nil {
			return fmt.Errorf("failed to verify if origin is registered: %w", err)
		}
		if !registered {
			return fiber.NewError(fiber.StatusBadRequest, "origin not registered")
		}

		return c.Next()
	}
}

package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
)

type NonRegisteredOriginFilter fiber.Handler

// ProvideNonRegisteredOriginFilter is wire provider for middleware that filter
// request with non registered origins.
func ProvideNonRegisteredOriginFilter(originRegistry originregistry.Service) NonRegisteredOriginFilter {
	return func(c *fiber.Ctx) error {
		origin := utils.UnsafeString(c.Request().Header.Peek(fiber.HeaderOrigin))
		origin, found := strings.CutPrefix(origin, "https://")
		if !found {
			origin = strings.TrimPrefix(origin, "http://")
		}

		registered, err := originRegistry.IsOriginRegistered(c.UserContext(), origin)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if !registered {
			return fiber.NewError(fiber.StatusBadRequest, "origin not registered")
		}

		return c.Next()
	}
}

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/services/orgs"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
)

type GetIndex fiber.Handler

func ProvideGetIndex(orgsService orgs.Service) GetIndex {
	return func(c *fiber.Ctx) error {
		userSession := c.Locals(middlewares.SessionKey{}).(sessions.Session)

		orgs, err := orgsService.ListOrgs(c.UserContext(), userSession.UserId())
		if err != nil {
			mustRender(c, fiber.StatusInternalServerError,
				"index", fiber.Map{
					"error": "Internal server error, please try again later",
				},
			)
			return err
		}

		return c.Render("index", fiber.Map{"orgs": orgs})
	}
}

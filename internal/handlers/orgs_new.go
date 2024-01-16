package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/services/orgs"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
)

type GetOrgsNew fiber.Handler

func ProvideGetOrgsNew() GetOrgsNew {
	return func(c *fiber.Ctx) error {
		return c.Render("orgs_new", fiber.Map{})
	}
}

type PostOrgsNew fiber.Handler

func ProvidePostOrgsNew(orgsService orgs.Service) PostOrgsNew {
	return func(c *fiber.Ctx) error {
		userSession := c.Locals(middlewares.SessionKey{}).(sessions.Session)

		type request struct {
			Name string `form:"name"`
		}

		req := request{}
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		// Validate org name.
		orgName, err := orgs.NewOrgName(req.Name)
		if err != nil {
			mustRender(c, fiber.StatusBadRequest,
				"orgs_new", fiber.Map{
					"error": err.Error(),
				},
			)
			return err
		}

		// Create organization.
		org, err := orgsService.CreateOrg(c.UserContext(), userSession.UserId(), orgName)
		if err != nil {
			if errors.Is(err, orgs.ErrOrgNameAlreadyTaken) {
				mustRender(c, fiber.StatusBadRequest,
					"orgs_new", fiber.Map{
						"error": err.Error(),
					},
				)

				return nil
			}

			mustRender(c, fiber.StatusBadRequest,
				"orgs_new", fiber.Map{
					"error": "Internal server error, please try again later",
				},
			)
			return err
		}

		return c.Redirect("/orgs/" + org.Id.String())
	}
}

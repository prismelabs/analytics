package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/services/orgs"
)

type GetOrgsOrgId fiber.Handler

// ProvideGetOrgsOrgId is a wire provider for get /orgs/:org_id handler.
func ProvideGetOrgsOrgId(orgsService orgs.Service) GetOrgsOrgId {
	return func(c *fiber.Ctx) error {
		rawOrgId := c.Params("org_id")
		orgId, err := orgs.ParseOrgId(rawOrgId)
		if err != nil {
			return c.Redirect("/")
		}

		org, err := orgsService.GetOrgById(c.UserContext(), orgId)
		if err != nil {
			if errors.Is(err, orgs.ErrOrgNotFound) {
				return fiber.ErrNotFound
			}

			mustRender(c, fiber.StatusInternalServerError,
				"orgs_orgid",
				fiber.Map{
					"error": "Internal server error, please try again later",
				},
			)
			return err
		}

		return c.Render("orgs_orgid", fiber.Map{
			"org": org,
		})
	}
}

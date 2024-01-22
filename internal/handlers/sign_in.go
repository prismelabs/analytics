package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/services/auth"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

type GetSignIn fiber.Handler

// ProvidePostSignUp define a wire provider for GET sign in handler.
func ProvideGetSignIn() GetSignIn {
	return func(c *fiber.Ctx) error {
		return c.Render("sign_in", fiber.Map{})
	}
}

type PostSignIn fiber.Handler

// ProvidePostSignUp define a wire provider for POST sign in handler.
func ProvidePostSignIn(authService auth.Service, sessionsService sessions.Service) PostSignIn {
	return func(c *fiber.Ctx) error {
		type request struct {
			Email    string `form:"email"`
			Password string `form:"password"`
		}

		req := request{}
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		// Validate request.

		// Validate email.
		email, err := users.NewEmail(req.Email)
		if err != nil {
			mustRender(c, fiber.StatusBadRequest,
				"sign_in", fiber.Map{
					"error": err.Error(),
				},
			)
			return err
		}

		password := secret.New(req.Password)

		user, err := authService.AuthenticateByPassword(c.UserContext(), email, password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				mustRender(c, fiber.StatusUnauthorized,
					"sign_in", fiber.Map{
						"error": err.Error(),
					},
				)
				return nil
			}

			mustRender(c, fiber.StatusInternalServerError,
				"sign_in", fiber.Map{
					"error": "Internal server error, please try again later",
				},
			)
			return err
		}

		err = sessionsService.CreateSession(c, user.Id)
		if err != nil {
			mustRender(c, fiber.StatusInternalServerError,
				"sign_in", fiber.Map{
					"error": "Internal server error, please try again later",
				},
			)
		}

		return c.Redirect("/")
	}
}

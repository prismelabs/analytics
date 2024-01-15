package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

type GetSignUp fiber.Handler

// ProvideGetSignUp define a wire provider for GET sign up handler.
func ProvideGetSignUp() GetSignUp {
	return func(c *fiber.Ctx) error {
		return c.Render("sign_up", fiber.Map{
			"title": "Sign up - Prisme Analytics",
		})
	}
}

type PostSignUp fiber.Handler

// ProvidePostSignUp define a wire provider for POST sign up handler.
func ProvidePostSignUp(userService users.Service) PostSignUp {
	return func(c *fiber.Ctx) error {
		type request struct {
			Name     string `form:"name"`
			Email    string `form:"email"`
			Password string `form:"password"`
		}

		req := request{}
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		// Validate request.

		// Validate user name.
		userName, err := users.NewUserName(req.Name)
		if err != nil {
			mustRender(c, fiber.StatusBadRequest,
				"sign_up", fiber.Map{
					"title": "Sign up - Prisme Analytics",
					"error": err.Error(),
				},
			)
			return err
		}

		// Validate email.
		email, err := users.NewEmail(req.Email)
		if err != nil {
			mustRender(c, fiber.StatusBadRequest,
				"sign_up", fiber.Map{
					"title": "Sign up - Prisme Analytics",
					"error": err.Error(),
				},
			)
			return err
		}

		// Validate password.
		password, err := users.NewPassword(secret.New(req.Password))
		if err != nil {
			mustRender(c, fiber.StatusBadRequest,
				"sign_up", fiber.Map{
					"title": "Sign up - Prisme Analytics",
					"error": err.Error(),
				},
			)
			return err
		}

		// Create user.
		_, err = userService.CreateUser(c.UserContext(), users.CreateCmd{
			UserName: userName,
			Email:    email,
			Password: password,
		})
		if err != nil {
			if errors.Is(err, users.ErrUserAlreadyExists) {
				mustRender(c, fiber.StatusBadRequest,
					"sign_up", fiber.Map{
						"title": "Sign up - Prisme Analytics",
						"error": "email already taken",
					},
				)
				return nil
			}

			mustRender(c, fiber.StatusInternalServerError,
				"sign_up",
				fiber.Map{
					"title": "Sign up - Prisme Analytics",
					"error": "Internal server error, please try again later",
				},
			)
			return err
		}

		return c.Redirect("/")
	}
}

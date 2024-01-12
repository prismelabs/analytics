package handlers

import "github.com/gofiber/fiber/v2"

// mustRender render template and sets the given status code. If an error occured
// this function panic.
func mustRender(c *fiber.Ctx, statusCode int, name string, bind interface{}, layouts ...string) {
	c.Response().SetStatusCode(statusCode)
	err := c.Render(name, bind, layouts...)
	if err != nil {
		panic(err)
	}
}

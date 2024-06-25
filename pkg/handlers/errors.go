package handlers

import (
	"github.com/gofiber/fiber/v2"
)

var (
	errSessionNotFound = fiber.NewError(fiber.StatusBadRequest, "session not found")
)

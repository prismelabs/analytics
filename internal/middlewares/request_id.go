package middlewares

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
)

const RequestIdKey = "request-id"

func RequestId(cfg config.Server) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var reqId string

			if cfg.TrustProxy {
				reqId = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			if reqId == "" {
				reqId = uuid.New().String()
			}

			c.Set(RequestIdKey, reqId)

			return next(c)
		}
	}
}

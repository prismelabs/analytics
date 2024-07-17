package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/rs/zerolog"
)

type GetNoscriptEventsIdentify fiber.Handler

// ProvideGetNoscriptEventsIdentify is a wire provider for
// GET /api/v1/noscript/events/custom/:name handler.
func ProvideGetNoscriptEventsIdentify(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) GetNoscriptEventsIdentify {
	return func(c *fiber.Ctx) error {
		err := c.Send(embedded.NoscriptGif)
		if err != nil {
			return err
		}

		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		return eventIdentifyHandler(
			c.UserContext(),
			logger,
			eventStore,
			saltManagerService,
			sessionStorage,
			requestReferrer,
			c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()),
			dataview.FasthttpArgsKvView{Args: c.Context().QueryArgs()},
			dataview.FasthttpArgsKeysValuesCollector{
				Args:           c.Context().QueryArgs(),
				Prefix:         "set-",
				ValueValidator: dataview.JsonValidator,
			},
			dataview.FasthttpArgsKeysValuesCollector{
				Args:           c.Context().QueryArgs(),
				Prefix:         "set-once-",
				ValueValidator: dataview.JsonValidator,
			},
		)
	}
}

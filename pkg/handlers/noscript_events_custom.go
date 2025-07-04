package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/embedded"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
)

type GetNoscriptEventsCustom fiber.Handler

// ProvideGetNoscriptEventsCustom is a wire provider for
// GET /api/v1/noscript/events/custom/:name handler.
func ProvideGetNoscriptEventsCustom(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstore.Service,
) GetNoscriptEventsCustom {
	return func(c *fiber.Ctx) error {
		err := c.Send(embedded.NoscriptGif)
		if err != nil {
			return err
		}

		// Referrer of the GET request, that is the viewed page.
		requestReferrer, err := hutils.PeekAndParseReferrerQueryHeader(c)
		if err != nil {
			return err
		}

		return eventsCustomHandler(
			c.UserContext(),
			eventStore,
			saltManagerService,
			sessionStorage,
			requestReferrer,
			c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()),
			c.Params("name"),
			dataview.FasthttpArgsKeysValuesCollector{
				Args:           c.Context().QueryArgs(),
				Prefix:         "prop-",
				ValueValidator: dataview.JsonValidator,
			},
		)
	}
}

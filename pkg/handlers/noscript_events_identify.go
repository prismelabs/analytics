package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
)

type GetNoscriptEventsIdentify fiber.Handler

// ProvideGetNoscriptEventsIdentify is a wire provider for
// GET /api/v1/noscript/events/custom/:name handler.
func ProvideGetNoscriptEventsIdentify(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) GetNoscriptEventsIdentify {
	return func(c *fiber.Ctx) error {
		err := c.Send(embedded.NoscriptGif)
		if err != nil {
			return err
		}

		for i := 1; i <= 3; i++ {
			err = eventsCustomHandler(
				c.UserContext(),
				eventStore,
				saltManagerService,
				sessionStorage,
				peekReferrerHeader(c),
				c.Request().Header.UserAgent(),
				utils.UnsafeBytes(c.IP()),
				c.Params("name"),
				dataview.FasthttpArgsKeysValuesCollector{Args: c.Context().QueryArgs(), Prefix: "prop-"},
			)
			if err == nil {
				break
			}

			// Retry as session may not exist.
			if errors.Is(err, errSessionNotFound) {
				time.Sleep(time.Duration(i) * 100 * time.Millisecond)
				continue
			}

			return err
		}

		return err
	}
}

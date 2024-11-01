package handlers

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/event"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type PostEventsPageviews fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST /api/v1/events/pageviews handler.
func ProvidePostEventsPageViews(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsPageviews {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer, err := hutils.PeekAndParseReferrerHeader(c)
		if err != nil {
			return err
		}

		return eventsPageviewsHandler(
			c.UserContext(),
			eventStore,
			uaParserService,
			ipGeolocatorService,
			saltManagerService,
			sessionStorage,
			&c.Request().Header,
			requestReferrer,
			c.Request().Header.Peek("X-Prisme-Document-Referrer"),
			c.Context().UserAgent(),
			utils.UnsafeBytes(c.IP()),
			utils.UnsafeString(c.Request().Header.Peek("X-Prisme-Status")),
			utils.UnsafeString(c.Request().Header.Peek("X-Prisme-Visitor-Id")),
		)
	}
}

func eventsPageviewsHandler(
	ctx context.Context,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
	headers *fasthttp.RequestHeader,
	requestReferrer uri.Uri,
	documentReferrer, userAgent, ipAddr []byte,
	status, visitorId string,
) (err error) {
	var referrerUri event.ReferrerUri
	pageView := event.PageView{
		PageUri: requestReferrer,
		Status:  fiber.StatusOK,
	}

	// Retrive pageview status code.
	if status != "" {
		pvStatus, err := strconv.ParseUint(status, 10, 16)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid pageview status")
		}
		pageView.Status = uint16(pvStatus)
	}

	// Parse referrer URI.
	referrerUri, err = event.ParseReferrerUri(documentReferrer)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid document referrer")
	}

	// Compute device id.
	deviceId := hutils.ComputeDeviceId(
		saltManagerService.StaticSalt().Bytes(), userAgent,
		ipAddr, utils.UnsafeBytes(pageView.PageUri.Host()),
	)

	isInternalTraffic := referrerUri.IsValid() && referrerUri.Host() == pageView.PageUri.Host()
	newSession := !isInternalTraffic

	// Internal traffic, session may already exists.
	if isInternalTraffic {
		var sessionExists bool
		// Increment pageview count.
		pageView.Session, sessionExists = sessionStorage.AddPageview(deviceId, referrerUri, pageView.PageUri)

		if !sessionExists {
			// Session with the given referrer URI doesn't exists but ones with
			// current page URI exists. This can happen if user refresh pages or a tab
			// is duplicated.
			// In both cases we want to insert a new session in temporary storage (memory)
			// but we don't want to send it to the persisten store (eventstore). This new
			// duplicated session will be persisted on next page view. This way we're sure
			// that duplicated session is used.
			{
				session, found := sessionStorage.WaitSession(deviceId, pageView.PageUri, time.Duration(0))
				if found {
					var err error
					session.SessionUuid, err = uuid.NewV7()
					if err != nil {
						return fmt.Errorf("failed to generate session uuid: %w", err)
					}

					session.PageUri = pageView.PageUri
					sessionStorage.InsertSession(deviceId, session)

					// Early return as we don't send event to the eventstore.
					return nil
				}
			}

			// Otherwise, simply create a new session.
			newSession = true
		} else {
			pageView.Timestamp = time.Now().UTC()

			// Update session visitor ID if needed.
			if visitorId != "" && pageView.Session.VisitorId != visitorId {
				visitorId = utils.CopyString(visitorId)
				pageView.Session, sessionExists = sessionStorage.IdentifySession(deviceId, pageView.PageUri, visitorId)
				if !sessionExists {
					return fmt.Errorf("failed to identify session after adding page view, this is probably a Prisme bug")
				}
			}
		}
	}

	// Create session.
	if newSession {
		// Filter bot.
		client := uaParserService.ParseUserAgent(
			utils.UnsafeString(userAgent),
		)
		if client.IsBot {
			return fiber.NewError(fiber.StatusBadRequest, "bot session filtered")
		}
		hutils.ExtractClientHints(headers, &client)

		sessionUuid, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate session uuid: %w", err)
		}

		if !isInternalTraffic {
			session, found := sessionStorage.WaitSession(deviceId, pageView.PageUri, time.Duration(0))
			if found {
				session.SessionUuid = sessionUuid
				sessionStorage.InsertSession(deviceId, session)

				// Early return as we don't send event to the eventstore.
				return nil
			}
		}

		// Compute visitor id if none was provided along request.
		if visitorId == "" {
			visitorId = hutils.ComputeVisitorId(
				saltManagerService.DailySalt().Bytes(), userAgent,
				ipAddr, utils.UnsafeBytes(pageView.PageUri.Host()), binary.LittleEndian.AppendUint64(nil, deviceId),
			)
		} else {
			visitorId = utils.CopyString(visitorId)
		}

		// Parse page uri args.
		args := fasthttp.Args{}
		args.Parse(pageView.PageUri.QueryString())

		pageView.Session = event.Session{
			PageUri:       pageView.PageUri,
			ReferrerUri:   referrerUri,
			Client:        client,
			CountryCode:   ipGeolocatorService.FindCountryCodeForIP(utils.UnsafeString(ipAddr)),
			VisitorId:     visitorId,
			SessionUuid:   sessionUuid,
			Utm:           hutils.ExtractUtmParams(&args),
			PageviewCount: 1,
		}
		pageView.Timestamp = pageView.Session.SessionTime()

		sessionStorage.InsertSession(deviceId, pageView.Session)
	}

	// Store event.
	err = eventStore.StorePageView(ctx, &pageView)
	if err != nil {
		return fmt.Errorf("failed to store pageview event: %w", err)
	}

	return nil
}

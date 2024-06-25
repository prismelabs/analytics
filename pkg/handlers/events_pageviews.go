package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type PostEventsPageview fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST /api/v1/events/pageviews events handler.
func ProvidePostEventsPageViews(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		return eventsPageviewsHandler(
			c.UserContext(),
			logger,
			eventStore,
			uaParserService,
			ipGeolocatorService,
			saltManagerService,
			sessionStorage,
			c.Context().UserAgent(),
			c.Request().Header.Peek("X-Prisme-Document-Referrer"),
			requestReferrer,
			utils.UnsafeBytes(c.IP()),
			utils.UnsafeString(c.Request().Header.Peek("X-Prisme-Visitor-Id")),
		)
	}
}

func eventsPageviewsHandler(
	ctx context.Context,
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
	userAgent, documentReferrer, requestReferrer, ipAddr []byte,
	visitorId string,
) (err error) {
	referrerUri := event.ReferrerUri{}
	pageView := event.PageView{}

	// Parse referrer URI.
	referrerUri, err = event.ParseReferrerUri(documentReferrer)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid X-Prisme-Document-Referrer")
	}

	// Parse page URI.
	pageView.PageUri, err = uri.ParseBytes(requestReferrer)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
	}

	// Compute device id.
	deviceId := computeDeviceId(
		saltManagerService.StaticSalt().Bytes(), userAgent,
		ipAddr, utils.UnsafeBytes(pageView.PageUri.Host()),
	)

	isExternalReferrer := referrerUri.Host() != pageView.PageUri.Host()
	newSession := isExternalReferrer

	// Retrieve session.
	if !newSession {
		sessionExists := false
		if visitorId != "" {
			visitorId = utils.CopyString(visitorId)
			_, sessionExists = sessionStorage.IdentifySession(deviceId, visitorId)
		} else {
			_, sessionExists = sessionStorage.GetSession(deviceId)
		}

		// Session not found.
		// This can happen if tracking script is not installed on all pages,
		// or prisme instance was restarted.
		if !sessionExists {
			newSession = true
		} else {
			pageView.Session, sessionExists = sessionStorage.IncSessionPageviewCount(deviceId)
			if !sessionExists {
				logger.Panic().Msg("failed to increment session pageview count after GetSession returned a session: session not found")
			}
			pageView.Timestamp = time.Now().UTC()
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

		sessionUuid, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate session uuid: %w", err)
		}

		// Compute visitor id if none was provided along request.
		if visitorId == "" {
			visitorId = computeVisitorId("prisme_",
				saltManagerService.DailySalt().Bytes(), userAgent,
				ipAddr, utils.UnsafeBytes(pageView.PageUri.Host()), []byte(deviceId),
			)
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
			Utm:           extractUtmParams(&args),
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

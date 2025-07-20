package eventstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/retry"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

type clickhouseBackend struct {
	ch           clickhouse.Ch
	eventBatches [maxEventKind]driver.Batch
}

func init() {
	backendsFactory["clickhouse"] = newClickhouseBackend
}

func newClickhouseBackend(
	logger log.Logger,
	cfg any,
	source source.Driver,
	teardown teardown.Service,
) backend {
	ch := clickhouse.NewCh(logger, cfg.(clickhouse.Config), source, teardown)

	return &clickhouseBackend{
		ch:           ch,
		eventBatches: [maxEventKind]driver.Batch{},
	}
}

// prepareBatch implements backend.
func (cb *clickhouseBackend) prepareBatch() error {
	queries := [maxEventKind]string{
		// pageviews table is a materialized view derived from sessions.
		// sessions table engine is VersionedCollapsedMergeTree so we can
		// keep appending row with the same Session UUID.
		// See https://clickhouse.com/docs/en/engines/table-engines/mergetree-family/versionedcollapsingmergetree
		pageviewEventKind:          "INSERT INTO sessions",
		customEventKind:            "INSERT INTO events_custom",
		fileDownloadEventKind:      "INSERT INTO file_downloads",
		outboundLinkClickEventKind: "INSERT INTO outbound_link_clicks",
	}

	for i := 0; i < len(cb.eventBatches); i++ {
		err := retry.LinearBackoff(3, time.Second, func(_ uint) error {
			var err error
			cb.eventBatches[i], err = cb.ch.PrepareBatch(context.Background(), queries[i])
			if err != nil {
				return err
			}
			return nil
		}, retry.NeverCancel)
		if err != nil {
			return err
		}
	}

	return nil
}

// appendToBatch implements backend.
func (cb *clickhouseBackend) appendToBatch(ev any) (err error) {
	switch e := ev.(type) {
	case *event.PageView:
		batch := cb.eventBatches[pageviewEventKind]

		if e.Session.PageviewCount > 1 {
			// Cancel previous session.
			err = batch.Append(
				e.Session.PageUri.Host(),
				e.Session.PageUri.Path(),
				e.Timestamp.UTC(),
				e.PageUri.Path(),
				e.Session.VisitorId,
				e.Session.SessionUuid,
				e.Session.Client.OperatingSystem,
				e.Session.Client.BrowserFamily,
				e.Session.Client.Device,
				e.Session.ReferrerUri.HostOrDirect(),
				e.Session.CountryCode,
				e.Session.Utm.Source,
				e.Session.Utm.Medium,
				e.Session.Utm.Campaign,
				e.Session.Utm.Term,
				e.Session.Utm.Content,
				e.Status,
				e.Session.PageviewCount-1, // Cancel previous version.
				-1,
			)
			if err != nil {
				return err
			}
		}

		return batch.Append(
			e.Session.PageUri.Host(),
			e.Session.PageUri.Path(),
			e.Timestamp.UTC(),
			e.PageUri.Path(),
			e.Session.VisitorId,
			e.Session.SessionUuid,
			e.Session.Client.OperatingSystem,
			e.Session.Client.BrowserFamily,
			e.Session.Client.Device,
			e.Session.ReferrerUri.HostOrDirect(),
			e.Session.CountryCode,
			e.Session.Utm.Source,
			e.Session.Utm.Medium,
			e.Session.Utm.Campaign,
			e.Session.Utm.Term,
			e.Session.Utm.Content,
			e.Status,
			e.Session.PageviewCount,
			1,
		)
	case *event.Custom:
		batch := cb.eventBatches[customEventKind]
		return batch.Append(
			e.Timestamp.UTC(),
			e.Session.PageUri.Host(),
			e.PageUri.Path(),
			e.Session.VisitorId,
			e.Session.SessionUuid,
			e.Name,
			e.Keys,
			e.Values,
		)

	case *event.OutboundLinkClick:
		batch := cb.eventBatches[outboundLinkClickEventKind]
		return batch.Append(
			e.Timestamp.UTC(),
			e.Session.PageUri.Host(),
			e.PageUri.Path(),
			e.Session.VisitorId,
			e.Session.SessionUuid,
			e.Link,
		)

	case *event.FileDownload:
		batch := cb.eventBatches[fileDownloadEventKind]
		return batch.Append(
			e.Timestamp.UTC(),
			e.Session.PageUri.Host(),
			e.PageUri.Path(),
			e.Session.VisitorId,
			e.Session.SessionUuid,
			e.FileUrl,
		)

	default:
		panic(fmt.Errorf("unknown event kind: %T", ev))
	}
}

// sendBatch implements backend.
func (cb *clickhouseBackend) sendBatch() error {
	var errs [maxEventKind]error
	ch := make(chan error)

	for i := 0; i < int(maxEventKind); i++ {
		batch := cb.eventBatches[i]
		go func() {
			ch <- retry.LinearBackoff(3, time.Second, func(_ uint) error {
				return batch.Send()
			}, retry.NeverCancel)
		}()
	}

	for i := 0; i < int(maxEventKind); i++ {
		errs[i] = <-ch
	}

	return errors.Join(errs[:]...)
}

func (cb *clickhouseBackend) query(ctx context.Context, query string, args ...any) (QueryResult, error) {
	return cb.ch.Query(ctx, query, args...)
}

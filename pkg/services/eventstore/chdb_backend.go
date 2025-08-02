//go:build chdb

package eventstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

func init() {
	backendsFactory["chdb"] = newChDbBackend
}

type chdbBackend struct {
	chdb         chdb.ChDb
	eventBatches [maxEventKind]*batch
}

func newChDbBackend(db eventdb.Service, teardown teardown.Service) backend {
	b := &chdbBackend{
		chdb:         db.Driver().(chdb.ChDb),
		eventBatches: [maxEventKind]*batch{},
	}

	teardown.RegisterProcedure(func() error {
		workDir := path.Join(os.TempDir(), "prisme-"+strconv.Itoa(os.Getpid()))
		return os.RemoveAll(workDir)
	})

	return b
}

var (
	eventKindTable = [maxEventKind]string{
		// pageviews table is a materialized view derived from sessions.
		// sessions table engine is VersionedCollapsedMergeTree so we can
		// keep appending row with the same Session UUID.
		// See https://clickhouse.com/docs/en/engines/table-engines/mergetree-family/versionedcollapsingmergetree
		pageviewEventKind:          "sessions_versionned",
		customEventKind:            "events_custom",
		fileDownloadEventKind:      "file_downloads",
		outboundLinkClickEventKind: "outbound_link_clicks",
	}
)

func (cb *chdbBackend) workDir() string {
	return path.Join(os.TempDir(), "prisme-"+strconv.Itoa(os.Getpid()), "eventstore")
}

// prepareBatch implements backend.
func (cb *chdbBackend) prepareBatch() error {
	workDir := cb.workDir()

	err := os.MkdirAll(workDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create event store work directory %s: %w", workDir, err)
	}

	for i := range int(maxEventKind) {
		filePath := path.Join(workDir, eventKindTable[i])
		cb.eventBatches[i], err = newBatch(cb.chdb, filePath, "insert into "+eventKindTable[i]+" from infile '"+filePath+"' FORMAT JSONEachRow")
		if err != nil {
			return err
		}
	}

	return nil
}

// appendToBatch implements backend.
func (cb *chdbBackend) appendToBatch(ev any) error {
	switch e := ev.(type) {
	case *event.PageView:
		tab := cb.eventBatches[pageviewEventKind]

		if e.Session.PageviewCount > 1 {
			err := tab.append(pageview{
				Domain:          e.Session.PageUri.Host(),
				EntryPath:       e.Session.PageUri.Path(),
				ExitTimestamp:   e.Timestamp.UTC().Format(time.DateTime),
				ExitPath:        e.PageUri.Path(),
				VisitorId:       e.Session.VisitorId,
				SessionUuid:     e.Session.SessionUuid,
				OperatingSystem: e.Session.Client.OperatingSystem,
				BrowserFamily:   e.Session.Client.BrowserFamily,
				Device:          e.Session.Client.Device,
				ReferrerDomain:  e.Session.ReferrerUri.HostOrDirect(),
				CountryCode:     e.Session.CountryCode.String(),
				UtmSource:       e.Session.Utm.Source,
				UtmMedium:       e.Session.Utm.Medium,
				UtmCampaign:     e.Session.Utm.Campaign,
				UtmTerm:         e.Session.Utm.Term,
				UtmContent:      e.Session.Utm.Content,
				Version:         e.Status,
				ExitStatus:      e.Session.PageviewCount - 1, // Cancel previous version.
				Sign:            -1,
			})
			if err != nil {
				return err
			}
		}

		return tab.append(pageview{
			Domain:          e.Session.PageUri.Host(),
			EntryPath:       e.Session.PageUri.Path(),
			ExitTimestamp:   e.Timestamp.UTC().Format(time.DateTime),
			ExitPath:        e.PageUri.Path(),
			VisitorId:       e.Session.VisitorId,
			SessionUuid:     e.Session.SessionUuid,
			OperatingSystem: e.Session.Client.OperatingSystem,
			BrowserFamily:   e.Session.Client.BrowserFamily,
			Device:          e.Session.Client.Device,
			ReferrerDomain:  e.Session.ReferrerUri.HostOrDirect(),
			CountryCode:     e.Session.CountryCode.String(),
			UtmSource:       e.Session.Utm.Source,
			UtmMedium:       e.Session.Utm.Medium,
			UtmCampaign:     e.Session.Utm.Campaign,
			UtmTerm:         e.Session.Utm.Term,
			UtmContent:      e.Session.Utm.Content,
			Version:         e.Status,
			ExitStatus:      e.Session.PageviewCount,
			Sign:            1,
		})

	case *event.Custom:
		tab := cb.eventBatches[customEventKind]
		return tab.append(customEvent{
			Timestamp:   e.Timestamp.UTC().Format(time.DateTime),
			Domain:      "",
			Path:        e.Session.PageUri.Path(),
			VisitorId:   e.Session.VisitorId,
			SessionUuid: e.Session.SessionUuid,
			Name:        e.Name,
			Keys:        e.Keys,
			Values:      e.Values,
		})

	case *event.OutboundLinkClick:
		tab := cb.eventBatches[outboundLinkClickEventKind]
		return tab.append(outboundLinkClick{
			Timestamp:   e.Timestamp.UTC().Format(time.DateTime),
			Domain:      e.Session.PageUri.Host(),
			Path:        e.Session.PageUri.Path(),
			VisitorId:   e.Session.VisitorId,
			SessionUuid: e.Session.SessionUuid,
			Link:        e.Link.String(),
		})

	case *event.FileDownload:
		tab := cb.eventBatches[fileDownloadEventKind]
		return tab.append(fileDownload{
			Timestamp:   e.Timestamp.UTC().Format(time.DateTime),
			Domain:      e.Session.PageUri.Host(),
			Path:        e.Session.PageUri.Path(),
			VisitorId:   e.Session.VisitorId,
			SessionUuid: e.Session.SessionUuid,
			FileUrl:     e.FileUrl.String(),
		})

	default:
		panic(fmt.Errorf("unknown event kind: %T", ev))
	}
}

// sendBatch implements backend.
func (cb *chdbBackend) sendBatch() error {
	var errs [maxEventKind]error
	ch := make(chan error)

	for _, b := range cb.eventBatches {
		go func(b *batch) {
			ch <- b.send()
		}(b)
	}

	for i := 0; i < int(maxEventKind); i++ {
		errs[i] = <-ch
	}

	return errors.Join(errs[:]...)
}

type batch struct {
	chdb    chdb.ChDb
	query   string
	file    *os.File
	encoder *json.Encoder
}

func newBatch(chdb chdb.ChDb, filePath, query string) (*batch, error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	return &batch{
		chdb:    chdb,
		query:   query,
		file:    f,
		encoder: json.NewEncoder(f),
	}, nil
}

func (b *batch) append(ev any) error {
	return b.encoder.Encode(ev)
}

func (b *batch) send() error {
	closeErr := b.file.Close()
	_, insertErr := b.chdb.Exec(b.query)
	return errors.Join(closeErr, insertErr)
}

type pageview struct {
	Domain          string    `json:"domain"`
	EntryPath       string    `json:"entry_path"`
	ExitTimestamp   string    `json:"exit_timestamp"`
	ExitPath        string    `json:"exit_path"`
	VisitorId       string    `json:"visitor_id"`
	SessionUuid     uuid.UUID `json:"session_uuid"`
	OperatingSystem string    `json:"operating_system"`
	BrowserFamily   string    `json:"browser_family"`
	Device          string    `json:"device"`
	ReferrerDomain  string    `json:"referrer_domain"`
	CountryCode     string    `json:"country_code"`
	UtmSource       string    `json:"utm_source"`
	UtmMedium       string    `json:"utm_medium"`
	UtmCampaign     string    `json:"utm_campaign"`
	UtmTerm         string    `json:"utm_term"`
	UtmContent      string    `json:"utm_content"`
	Version         uint16    `json:"version"`
	ExitStatus      uint16    `json:"exit_status"`
	Sign            int       `json:"sign"`
}

type customEvent struct {
	Timestamp   string    `json:"timestamp"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	VisitorId   string    `json:"visitor_id"`
	SessionUuid uuid.UUID `json:"session_uuid"`
	Name        string    `json:"name"`
	Keys        []string  `json:"keys"`
	Values      []string  `json:"values"`
}

type outboundLinkClick struct {
	Timestamp   string    `json:"timestamp"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	VisitorId   string    `json:"visitor_id"`
	SessionUuid uuid.UUID `json:"session_uuid"`
	Link        string    `json:"link"`
}

type fileDownload struct {
	Timestamp   string    `json:"timestamp"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	VisitorId   string    `json:"visitor_id"`
	SessionUuid uuid.UUID `json:"session_uuid"`
	FileUrl     string    `json:"url"`
}

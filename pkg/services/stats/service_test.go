//go:build test && !race && chdb

package stats

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/testutils/faker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestIntegNoRaceDetectorService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var (
		cfg = eventstore.Config{
			MaxBatchSize:      1,
			MaxBatchTimeout:   time.Microsecond,
			RingBuffersFactor: 100,
		}
		db           eventdb.Service
		err          error
		promRegistry *prometheus.Registry
		stats        Service
		store        eventstore.Service
		ctx          = context.Background()
	)

	forEachEventStoreBackend := func(t *testing.T, test func(t *testing.T)) {
		eventdb.ForEachDriver(t, func(edb eventdb.Service) {
			promRegistry = prometheus.NewRegistry()
			teardown := teardown.NewService()
			db = edb

			_ = io.Discard
			t.Run(edb.DriverName(), func(t *testing.T) {
				store, err = eventstore.NewService(
					cfg,
					db,
					log.New("stats-test", io.Discard, false),
					promRegistry,
					teardown,
				)
				stats = NewService(db, teardown)
				require.NoError(t, err)
				test(t)
				require.NoError(t, teardown.Teardown())
			})
		})
	}

	t.Run("Visitors/Sessions/Pageviews", func(t *testing.T) {
		forEachEventStoreBackend(t, func(t *testing.T) {
			df, err := stats.Visitors(ctx, Filters{})
			require.NoError(t, err)
			require.Len(t, df.Keys, 0)

			df, err = stats.Sessions(ctx, Filters{})
			require.NoError(t, err)
			require.Len(t, df.Keys, 0)

			df, err = stats.PageViews(ctx, Filters{})
			require.NoError(t, err)
			require.Len(t, df.Keys, 0)

			df, err = stats.LiveVisitors(ctx, Filters{})
			require.NoError(t, err)
			require.Len(t, df.Keys, 0)

			df, err = stats.Bounces(ctx, Filters{})
			require.NoError(t, err)
			require.Len(t, df.Keys, 0)

			// Session 1.
			now := time.Now()
			session := faker.Session()
			visitorId := session.VisitorId
			session.SessionUuid = faker.UuidV7(now)
			session.PageviewCount++
			pv := faker.PageView(session)
			require.NoError(t, store.StorePageView(ctx, &pv))

			// Session 2.
			{
				session := faker.Session()
				session.VisitorId = visitorId
				session.SessionUuid = faker.UuidV7(now)
				for range 2 {
					session.PageviewCount++
					pv := faker.PageView(session)
					require.NoError(t, store.StorePageView(ctx, &pv))
				}
			}

			time.Sleep(time.Second)

			df, err = stats.Visitors(ctx, Filters{})
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(df.Keys), 1)
			require.EqualValues(t, 1, sum(df.Values))

			df, err = stats.Sessions(ctx, Filters{})
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(df.Keys), 1)
			require.EqualValues(t, 2, sum(df.Values))

			df, err = stats.PageViews(ctx, Filters{})
			require.NoError(t, err)
			require.EqualValues(t, 3, sum(df.Values))

			df, err = stats.LiveVisitors(ctx, Filters{})
			require.NoError(t, err)
			require.EqualValues(t, 1, sum(df.Values))

			df, err = stats.Bounces(ctx, Filters{})
			require.NoError(t, err)
			require.EqualValues(t, 1, sum(df.Values))
		})
	})
}

func sum(s []uint64) uint64 {
	var sum uint64 = 0
	for _, i := range s {
		sum += i
	}
	return sum
}

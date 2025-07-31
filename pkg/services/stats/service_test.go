package stats

// func TestIntegNoRaceDetectorService(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}
//
// 	for _, backend := range []string{"clickhouse", "chdb"} {
// 		t.Run(backend, func(t *testing.T) {
// 			logger := log.New("stats_service_test", io.Discard, true)
// 			teardownService := teardown.NewService()
// 			source := clickhouse.EmbeddedSourceDriver(logger)
// 			defer func() { require.NoError(t, teardownService.Teardown()) }()
//
// 			var backendCfg any
// 			switch backend {
// 			case "clickhouse":
// 				var cfg clickhouse.Config
// 				testutils.ConfigueLoad(t, &cfg)
// 				backendCfg = cfg
// 			case "chdb":
// 				var cfg chdb.Config
// 				testutils.ConfigueLoad(t, &cfg)
// 				backendCfg = cfg
// 			default:
// 				panic("unkown event store backend")
// 			}
//
// 			cfg := eventstore.Config{
// 				Backend:           backend,
// 				BackendConfig:     backendCfg,
// 				MaxBatchSize:      1,
// 				MaxBatchTimeout:   time.Millisecond,
// 				RingBuffersFactor: 100,
// 			}
//
// 			promRegistry := prometheus.NewRegistry()
// 			eventStore, err := eventstore.NewService(cfg, logger, promRegistry, teardownService, source)
// 			require.NoError(t, err)
//
// 			service := New(eventStore)
//
// 			now := time.Now()
// 			h24 := 24 * time.Hour
// 			ctx := context.Background()
//
// 			t.Run("Last24Hours", func(t *testing.T) {
// 				// Ingest data.
// 				eventStore.StorePageView(ctx, &event.PageView{
// 					Session: event.Session{
// 						PageUri:     testutils.Must(uri.Parse)("http://www.example.com/foo"),
// 						ReferrerUri: testutils.Must(event.ParseReferrerUri)([]byte("http://www.example.com/referrer")),
// 						Client: uaparser.Client{
// 							BrowserFamily:   "Firefox",
// 							OperatingSystem: "macOS",
// 							Device:          "Unkown",
// 							IsBot:           false,
// 						},
// 						CountryCode:   ipgeolocator.CountryCode{},
// 						VisitorId:     "XXX",
// 						SessionUuid:   testutils.MustNoArg(uuid.NewV7)(),
// 						Utm:           event.UtmParams{},
// 						PageviewCount: 1,
// 					},
// 					Timestamp: now.Add(-time.Minute),
// 					PageUri:   testutils.Must(uri.Parse)("http://www.example.com/foo"),
// 					Status:    200,
// 				})
//
// 				b := service.Begin(ctx, TimeRange{
// 					Start: now.Add(-h24),
// 					Dur:   h24,
// 				}, Filters{})
//
// 				sessions, err := b.Sessions()
// 				require.NoError(t, err)
// 				_ = sessions
// 			})
// 		})
// 	}
// }

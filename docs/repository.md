# Repository structure

This document explains repository structure convention.

```
$ tree -d --gitignore .
.
├── cmd
│   ├── addevents                 # CLI utils to add millions fictive events to clickhouse and test clickhouse ingestion performance
│   ├── server                    # Prisme server main.go
│   └── uaparser                  # CLI utils to generate test data.
├── config                        # Directory containing scripts to generate .env config file.
├── deploy                        # Directory containing deployment files.
├── docs                          # Documentation.
├── mocks                         # Mocks holds mocks website used to test manually Prisme when developping.
├── pkg                           # Prisme Go packages.
├── tests                         # End-to-end tests.
│   ├── bun                       # E2E tests using bun.
│   └── perf                      # E2E benchmarks.
├── tracker                       # Vanilla JS trackers scripts.
```

Note that each directory contains a `README.md` documenting their purposes and
what they should contains. Go packages contains a `doc.go` file instead so
documentation can be read using `godoc -http=:6060`

## `pkg` - Go packages

This section details convention about `pkg/` structure. It isn't a documentation
of all packages.

```
pkg
├── embedded                      # Files embedded within Go binary.
├── event                         # Prisme events structure.
├── handlers                      # HTTP handlers.
├── middlewares                   # HTTP middlewares.
├── services                      # Reusable and swappable features hided behind interfaces.
└── wired                         # Wire providers of external dependency.
```


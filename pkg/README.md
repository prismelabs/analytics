# `pkg/` - Go packages

This document details convention about `pkg/` structure. Packages documentation
is available using `godoc -http=:6060`.

```
pkg
├── chdb                          # chdb initialization and migration.
├── clickhouse                    # ClickHouse initialization and migration.
├── dataview                      # Key-value view over structured data (query args, JSON object)
├── embedded                      # Files embedded within Go binary.
├── event                         # Events as dump structs.
├── handlers                      # HTTP handlers.
├── log                           # Structured logging.
├── middlewares                   # HTTP middlewares.
├── options                       # Glue code and data for configuration.
├── retry                         # Linear, random and exponential retry helpers.
├── services                      # Reusable and swappable features hidden behind interfaces.
├── sql                           # Fast & dumb SQL builder.
├── testutils                     # Helper functions for writing tests and benchmarks.
└── uri                           # A zero-copy URI parser and type.
```


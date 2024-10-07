# `pkg/` - Go packages

This document details convention about `pkg/` structure. Packages documentation
is available using `godoc -http=:6060`.

```
pkg
├── embedded                      # Files embedded within Go binary.
├── event                         # Prisme events structure.
├── handlers                      # HTTP handlers.
├── middlewares                   # HTTP middlewares.
├── services                      # Reusable and swappable features hided behind interfaces.
└── wired                         # Wire providers of external dependency.
```


package options

import "github.com/negrel/configue"

// Proxy options.
type Proxy struct {
	// Trust proxy headers.
	Trust bool
	// X-Forwarded-For proxy header.
	ForwardedForHeader string
	// X-Request-Id proxy header.
	RequestIdHeader string
}

// RegisterOptions registers options in provided Figue.
func (p *Proxy) RegisterOptions(f *configue.Figue) {
	f.BoolVar(&p.Trust, "proxy.trust", false, "trust proxy headers (X-Forwarded-For, X-Request-Id)")
	f.StringVar(&p.ForwardedForHeader, "proxy.header.xforwardedfor", "X-Forwarded-For", "HTTP header used to determine client IP address")
	f.StringVar(&p.RequestIdHeader, "proxy.header.requestid", "X-Request-Id", "HTTP header used to retrieve request ID")
}

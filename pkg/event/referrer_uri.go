package event

import (
	"github.com/prismelabs/analytics/pkg/uri"
)

// ReferrerUri wraps an uri.Uri to represent referrer URIs (absolute URI with
// no query args and fragment/hash). An empty referrer uri is considered as "direct".
type ReferrerUri struct {
	uri.Uri
}

// ParseReferrerUri parses a referrer uri.
func ParseReferrerUri(rawUri []byte) (ReferrerUri, error) {
	if len(rawUri) == 0 {
		return ReferrerUri{}, nil
	}

	uri, err := uri.ParseBytes(rawUri)
	return ReferrerUri{uri}, err
}

// HostOrDirect returns uri host or "direct" if uri is empty.
func (ru *ReferrerUri) HostOrDirect() string {
	if !ru.IsValid() {
		return "direct"
	}

	return ru.Host()
}

// String implements fmt.Stringer.
func (ru ReferrerUri) String() string {
	return ru.Scheme() + "://" + ru.Host() + ru.Path()
}

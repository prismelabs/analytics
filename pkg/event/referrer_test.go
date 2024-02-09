package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseReferrerDomain(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		invalidDomains := []string{
			"mydomain*com",
			"123domain!",
			"_invalid-domain.com",
			"space domain.com",
			"my domain .com",
			"domain#invalid.com",
			"-hyphenstart.com",
			"domain_with_underscores-.com",
		}

		for _, domain := range invalidDomains {
			t.Run(domain, func(t *testing.T) {
				referrerDomain, err := ParseReferrerDomain("http://" + domain + "/foo")
				require.Error(t, err)
				require.Equal(t, ReferrerDomain{}, referrerDomain)
			})
		}
	})

	t.Run("Valid", func(t *testing.T) {
		validDomains := []string{
			"alphabets123.com",
			"my-domain-name.com",
			"1234example.net",
			"tech-geeks.org",
			"secure-site.info",
			"bestblogsite.biz",
			"creative-web.dev",
			"xyz-company.co",
			"e-commerce-site.store",
			"travel-experts.travel",
			"xn--kn8h.to",
		}

		for _, domain := range validDomains {
			t.Run(domain, func(t *testing.T) {
				referrerDomain, err := ParseReferrerDomain("http://" + domain + "/foo")
				require.NoError(t, err)
				require.NotEqual(t, ReferrerDomain{}, referrerDomain)
				require.Equal(t, domain, referrerDomain.String())
			})
		}

		t.Run("🏹.to", func(t *testing.T) {
			url := "http://🏹.to/"
			referrer, err := ParseReferrerDomain(url)
			require.NoError(t, err)
			require.NotEqual(t, ReferrerDomain{}, referrer)
			require.Equal(t, "xn--kn8h.to", referrer.String())
		})

		t.Run("Direct", func(t *testing.T) {
			referrerDomain, err := ParseReferrerDomain("")
			require.NoError(t, err)
			require.Equal(t, "direct", referrerDomain.String())
		})
	})
}

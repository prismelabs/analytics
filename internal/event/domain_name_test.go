package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDomainName(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		invalidDomains := []string{
			"mydomain*com",
			"123domain!",
			"_invalid-domain.com",
			"space domain.com",
			"special@character.com",
			"my domain .com",
			"domain#invalid.com",
			"-hyphenstart.com",
			"domain_with_underscores-.com",
		}

		for _, domain := range invalidDomains {
			t.Run(domain, func(t *testing.T) {
				domainName, err := ParseDomainName(domain)
				require.Error(t, err)
				require.Equal(t, DomainName{}, domainName)
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
				domainName, err := ParseDomainName(domain)
				require.NoError(t, err)
				require.NotEqual(t, DomainName{}, domainName)
				require.Equal(t, domain, domainName.String())
			})
		}

		t.Run("🏹.to", func(t *testing.T) {
			domain := "🏹.to"
			domainName, err := ParseDomainName(domain)
			require.NoError(t, err)
			require.NotEqual(t, DomainName{}, domainName)
			require.Equal(t, "xn--kn8h.to", domainName.String())
		})
	})
}

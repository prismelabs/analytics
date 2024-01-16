package embedded

//go:embed pageview.js
var pageviewJs string

type PageviewJs string

// ProvidePageviewJs is a wire provider for pageview.js script.
func ProvidePageviewJs(cfg config.Server) PageviewJs {
	script := strings.ReplaceAll(pageviewJs, "{SERVER_URL}", cfg.Server.Url())

	return PageviewJs(script)
}

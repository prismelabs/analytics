package uaparser

import "regexp"

// See https://github.com/ua-parser/uap-core/blob/master/regexes.yaml
var (
	// Case insensitive regexes.
	isBotBrowserFamilyRegex = regexp.MustCompile(`(?i).*(Bot|Spider|Crawl|Headless).*`)
)

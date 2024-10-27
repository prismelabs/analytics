package embedded

import _ "embed"

//go:embed uap/regexes.yml
var UapRegexes []byte

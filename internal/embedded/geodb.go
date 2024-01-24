package embedded

import _ "embed"

//go:embed geodb/ip2asn-combined.mmdb
var Ip2AsnDb []byte

package ipgeolocator

import (
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"testing"

	"github.com/prismelabs/prismeanalytics/pkg/log"
)

func randIpv4Str() string {
	ip := rand.Uint32()
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, ip)

	return net.IP(buf).String()
}

func BenchmarkFindCountryCodeForIp(b *testing.B) {
	b.Run("IPv4", func(b *testing.B) {
		logger := log.NewLogger("ipgeolocator_mmdb_service", io.Discard, false)
		service := ProvideMmdbService(logger)

		for i := 0; i < b.N; i++ {
			_ = service.FindCountryCodeForIP(randIpv4Str())
		}
	})
}

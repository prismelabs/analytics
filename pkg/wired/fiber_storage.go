package wired

import (
	"time"

	"github.com/gofiber/storage"
	"github.com/gofiber/storage/memory"
)

// ProvideFiberStorage is a wire provider for fiber storage.
func ProvideFiberStorage() storage.Storage {
	return memory.New(memory.Config{
		GCInterval: 10 * time.Second,
	})
}

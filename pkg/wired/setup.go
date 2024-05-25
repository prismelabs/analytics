package wired

import "github.com/google/uuid"

type Setup struct{}

// ProvideSetup is a wire provider for setup.
func ProvideSetup() Setup {
	uuid.EnableRandPool()

	return Setup{}
}

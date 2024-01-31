package wired

type Setup struct{}

// ProvideSetup is a wire provider for setup.
func ProvideSetup() Setup { return Setup{} }

package prisme

import "github.com/rs/zerolog"

// App is a singleton use for Prisme global state.
type App struct {
	Config Config
	Logger zerolog.Logger
}

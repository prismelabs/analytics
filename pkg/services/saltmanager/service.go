package saltmanager

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// Service define a hashing salt manager.
type Service interface {
	// DailySalt returns today's salt.
	DailySalt() Salt
}

// ProvideService is a wire provider for hashing salt manager service.
func ProvideService(logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "saltmanager").
		Logger()

	srv := &service{
		logger: logger,
	}

	err := srv.rotateSalt()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to rotate initial salt")
	}

	go srv.rotateSaltLoop()

	return srv
}

type service struct {
	logger      zerolog.Logger
	currentSalt atomic.Pointer[Salt]
}

// DailySalt implements Service.
func (s *service) DailySalt() Salt {
	return *s.currentSalt.Load()
}

func (s *service) rotateSaltLoop() {
	for {
		// Tomorrow midnight.
		nextRotation := time.Now().AddDate(0, 0, 1)
		nextRotation = time.Date(nextRotation.Year(), nextRotation.Month(), nextRotation.Day(), 0, 0, 0, 0, time.UTC)

		time.Sleep(time.Until(nextRotation))

		err := s.rotateSalt()
		if err != nil {
			s.logger.Err(err).Msg("failed to rotate salt")
		}
	}
}

func (s *service) rotateSalt() error {
	// No current salt in database. Generate a new one.
	salt, err := randomSalt()
	if err != nil {
		return fmt.Errorf("failed to generate random salt: %w", err)
	}

	s.currentSalt.Store(&salt)
	return nil
}

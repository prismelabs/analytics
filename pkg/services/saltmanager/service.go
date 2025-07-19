package saltmanager

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/log"
)

// Service define a hashing salt manager.
type Service interface {
	// DailySalt returns today's salt.
	DailySalt() Salt
	// StaticSalt returns same salt until end of program.
	StaticSalt() Salt
}

// NewService returns a new hashing salt manager service.
func NewService(logger log.Logger) Service {
	logger = logger.With("service", "saltmanager")

	staticSalt, err := uuid.NewRandom()
	if err != nil {
		logger.Fatal("failed to generate static salt", err)
	}

	srv := &service{
		logger:     logger,
		staticSalt: Salt(staticSalt),
	}

	err = srv.rotateSalt()
	if err != nil {
		logger.Fatal("failed to rotate initial salt", err)
	}

	go srv.rotateSaltLoop()

	return srv
}

type service struct {
	logger      log.Logger
	currentSalt atomic.Pointer[Salt]
	staticSalt  Salt
}

// DailySalt implements Service.
func (s *service) DailySalt() Salt {
	return *s.currentSalt.Load()
}

// StaticSalt implements Service.
func (s *service) StaticSalt() Salt {
	return s.staticSalt
}

func (s *service) rotateSaltLoop() {
	for {
		// Tomorrow midnight.
		nextRotation := time.Now().AddDate(0, 0, 1)
		nextRotation = time.Date(nextRotation.Year(), nextRotation.Month(), nextRotation.Day(), 0, 0, 0, 0, time.UTC)

		time.Sleep(time.Until(nextRotation))

		err := s.rotateSalt()
		if err != nil {
			s.logger.Err("failed to rotate salt", err)
		} else {
			s.logger.Info("salt rotated")
		}
	}
}

func (s *service) rotateSalt() error {
	salt, err := randomSalt()
	if err != nil {
		return fmt.Errorf("failed to generate random salt: %w", err)
	}

	s.currentSalt.Store(&salt)
	return nil
}

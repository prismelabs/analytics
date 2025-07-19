package teardown

import (
	"errors"
)

// Procedure define a that can fail.
type Procedure func() error

// Service define a teardown procedure registry service.
type Service interface {
	RegisterProcedure(Procedure)
	Teardown() error
}

// NewService returns a new teardown procedure registry service.
func NewService() Service {
	return &service{
		procedures: nil,
	}
}

type service struct {
	procedures []Procedure
}

// RegisterProcedure implements Service.
func (s *service) RegisterProcedure(proc Procedure) {
	s.procedures = append(s.procedures, proc)
}

// Teardown implements Service.
func (s *service) Teardown() error {
	var finalErr error

	for i := 0; i < len(s.procedures); i++ {
		proc := s.procedures[len(s.procedures)-1-i]
		err := proc()
		if err != nil {
			finalErr = errors.Join(err)
		}
	}

	s.procedures = nil

	return finalErr
}

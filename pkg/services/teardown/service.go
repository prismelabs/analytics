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

// ProvideService is a wire provider for a teardown procedure registry service.
func ProvideService() Service {
	return &service{
		procedures: []Procedure{},
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

	for _, proc := range s.procedures {
		err := proc()
		if err != nil {
			finalErr = errors.Join(err)
		}
	}

	return finalErr
}

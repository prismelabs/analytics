package orgs

import (
	"context"
	"fmt"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

// Service define organizations management service.
type Service interface {
	CreateOrg(context.Context, users.UserId, OrgName) (Organization, error)
	ListOrgs(context.Context, users.UserId) ([]Organization, error)
	GetOrgById(context.Context, OrgId) (Organization, error)
}

// ProvideService is a wire provider for org management service.
func ProvideService(pg postgres.Pg) Service {
	return newService(pgStore{pg.DB})
}

func newService(store store) service {
	return service{store}
}

type service struct {
	store store
}

// CreateOrg implements Service.
func (s service) CreateOrg(ctx context.Context, userId users.UserId, orgName OrgName) (Organization, error) {
	org := Organization{
		Id:           NewOrgId(),
		OwnerId:      userId,
		GrafanaOrgId: 0, // TODO: create grafana org.
		Name:         orgName,
		CreatedAt:    time.Now(),
	}

	err := s.store.InsertOrg(ctx, org)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to insert org in store: %w", err)
	}

	return org, nil
}

// ListOrgs implements Service.
func (s service) ListOrgs(ctx context.Context, userId users.UserId) ([]Organization, error) {
	orgs, err := s.store.SelectOrgsByOwnerId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to select orgs by owner id from store: %w", err)
	}

	return orgs, nil
}

// GetOrgById implements Service.
func (s service) GetOrgById(ctx context.Context, orgId OrgId) (Organization, error) {
	// TODO: verify ownership ?

	org, err := s.store.SelectOrgById(ctx, orgId)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to select org by id from store: %w", err)
	}

	return org, nil
}

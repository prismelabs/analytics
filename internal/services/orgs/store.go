package orgs

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

var (
	ErrOrgNotFound         = errors.New("organization not found")
	ErrOrgNameAlreadyTaken = errors.New("organization name already taken")
)

//go:generate mockgen -source store.go -destination store_mock_test.go -package orgs -mock_names store=MockStore store
type store interface {
	InsertOrg(context.Context, Organization) error
	SelectOrgsByOwnerId(context.Context, users.UserId) ([]Organization, error)
	SelectOrgById(context.Context, OrgId) (Organization, error)
}

type pgStore struct {
	db *sql.DB
}

// InsertOrg implements store.
func (pgs pgStore) InsertOrg(ctx context.Context, org Organization) error {
	_, err := pgs.db.ExecContext(
		ctx,
		"INSERT INTO orgs(id, owner_id, name, grafana_org_id, created_at) VALUES ($1, $2, $3, $4, $5)",
		org.Id,
		org.OwnerId,
		org.Name,
		org.GrafanaOrgId,
		org.CreatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return ErrOrgNameAlreadyTaken
		}

		return err
	}

	return nil
}

// SelectOrgsByOwnerId implements store.
func (pgs pgStore) SelectOrgsByOwnerId(ctx context.Context, userId users.UserId) ([]Organization, error) {
	rows, err := pgs.db.QueryContext(
		ctx,
		"SELECT id, name, grafana_org_id, created_at FROM orgs WHERE owner_id = $1",
		userId,
	)
	if err != nil {
		return nil, err
	}

	var result []Organization

	// Scan rows.
	for rows.Next() {
		org := Organization{
			OwnerId: userId,
		}
		err := rows.Scan(&org.Id, &org.Name, &org.GrafanaOrgId, &org.CreatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, org)
	}

	return result, nil
}

// SelectOrgById implements store.
func (pgs pgStore) SelectOrgById(ctx context.Context, orgId OrgId) (Organization, error) {
	row := pgs.db.QueryRowContext(ctx,
		"SELECT owner_id, name, grafana_org_id, created_at FROM orgs WHERE id = $1",
		orgId,
	)

	org := Organization{
		Id: orgId,
	}

	err := row.Scan(&org.OwnerId, &org.Name, &org.GrafanaOrgId, &org.CreatedAt)
	if err != nil {
		return Organization{}, err
	}

	return org, nil
}

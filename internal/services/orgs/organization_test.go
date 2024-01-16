package orgs

import (
	"fmt"

	"go.uber.org/mock/gomock"
)

type OrganizationMatcher struct {
	Id           gomock.Matcher
	OwnerId      gomock.Matcher
	GrafanaOrgId gomock.Matcher
	Name         gomock.Matcher
	CreatedAt    gomock.Matcher
}

// Matches implements gomock.Matcher.
func (om OrganizationMatcher) Matches(x any) bool {
	if org, isOrg := x.(Organization); isOrg {
		if !om.Id.Matches(org.Id) {
			return false
		}
		if !om.OwnerId.Matches(org.OwnerId) {
			return false
		}
		if !om.GrafanaOrgId.Matches(org.GrafanaOrgId) {
			return false
		}
		if !om.Name.Matches(org.Name) {
			return false
		}
		if !om.CreatedAt.Matches(org.CreatedAt) {
			return false
		}

		return true
	}

	return false
}

// String implements gomock.Matcher.
func (om OrganizationMatcher) String() string {
	return fmt.Sprintf("{Id: %v, OwnerId: %v, GrafanaOrgId: %v, Name: %v, CreatedAt: %v}",
		om.Id, om.OwnerId, om.GrafanaOrgId, om.Name, om.CreatedAt)
}

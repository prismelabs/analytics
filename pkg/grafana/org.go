package grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

var (
	ErrGrafanaOrgAlreadyExists = errors.New("grafana org already exists")
	ErrGrafanaOrgNotFound      = errors.New("grafana org not found")
	ErrGrafanaUserAlreadyInOrg = errors.New("grafana user already in org")
)

// OrgId define a grafana organization id.
type OrgId int64

type OrgUser struct {
	User
	Role  Role  `json:"role"`
	OrgId OrgId `json:"orgId"`
}

// CreateOrg creates an organization with the given name.
// If it fails an error is returned. ErrGrafanaOrgAlreadyExists is returned
// if name is already used by an organization.
func (c Client) CreateOrg(ctx context.Context, name string) (OrgId, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(c.cfg.Url + "/api/orgs")

	type requestBody struct {
		Name string `json:"name"`
	}
	jsonBody, err := json.Marshal(requestBody{name})
	if err != nil {
		panic(err)
	}
	req.Header.SetContentType("application/json")
	req.SetBody(jsonBody)

	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to query grafana to create an organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() == 409 && strings.Contains(string(resp.Body()), "Organization name taken") {
		return 0, ErrGrafanaOrgAlreadyExists
	} else if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to create grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	type createOrgResp struct {
		OrgId OrgId `json:"orgId"`
	}
	respBody := createOrgResp{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return 0, fmt.Errorf("failed to parse grafana response: %w", err)
	}
	if respBody.OrgId == -1 {
		return 0, fmt.Errorf("invalid grafana response")
	}

	return respBody.OrgId, nil
}

// GetOrgByID retrieves organization name using it's id.
// If it fails an error is returned. ErrGrafanaOrgNotFound is returned
// if no organization has the given id.
func (c Client) GetOrgByID(ctx context.Context, orgId OrgId) (string, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v", c.cfg.Url, orgId))

	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return "", fmt.Errorf("failed to query grafana to get organization by id: %w", err)
	}

	// Grafana returns 403 forbidden when an organization doesn't exist.
	if resp.StatusCode() == 403 && strings.Contains(string(resp.Body()), "You'll need additional permissions to perform this action.") {
		return "", ErrGrafanaOrgNotFound
	} else if resp.StatusCode() != 200 {
		return "", fmt.Errorf("failed to get grafana organization by id: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	type lookupOrgResp struct {
		Name string `json:"name"`
	}
	respBody := lookupOrgResp{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse find grafana organization by id response: %w", err)
	}

	return respBody.Name, nil
}

// FindOrgByName retrieves organization id using it's name.
// If it fails an error is returned. ErrGrafanaOrgNotFound is returned
// if no organization has the given name.
func (c Client) FindOrgByName(ctx context.Context, name string) (OrgId, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(c.cfg.Url + "/api/orgs/name/" + name)

	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to query grafana to find an organization by name: %w", err)
	}

	if resp.StatusCode() == 404 {
		return 0, ErrGrafanaOrgNotFound
	} else if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to find grafana organization by name: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	type lookupOrgResp struct {
		Id OrgId `json:"id"`
	}
	respBody := lookupOrgResp{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return 0, fmt.Errorf("failed to parse find grafana organization by name response: %w", err)
	}

	return respBody.Id, nil
}

// GetOrCreateOrg retrieve organization id with the given name or create it.
func (c Client) GetOrCreateOrg(ctx context.Context, name string) (OrgId, error) {
	orgId, err := c.FindOrgByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrGrafanaOrgNotFound) {
			orgId, err = c.CreateOrg(ctx, name)
		} else {
			return 0, err
		}
	}

	return orgId, err
}

// UpdateOrgName updates organization name with the given id.
func (c Client) UpdateOrgName(ctx context.Context, orgId OrgId, name string) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("PUT")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v", c.cfg.Url, orgId))
	c.addAuthorizationHeader(req)

	type requestBody struct {
		Name string `json:"name"`
	}
	jsonBody, err := json.Marshal(requestBody{Name: name})
	if err != nil {
		panic(err)
	}
	req.SetBody(jsonBody)
	req.Header.SetContentType("application/json")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to update an organization name: %w", err)
	}

	if resp.StatusCode() == 404 {
		return ErrGrafanaOrgNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to update grafana organization name: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// AddUserToOrg adds the given user to the given organization with the given role.
// If user is already part of organization, ErrGrafanaUserAlreadyInOrg will be returned.
// If user is not found, ErrGrafanaUserNotFound will be returned.
// Any other error will return a generic error.
func (c Client) AddUserToOrg(ctx context.Context, orgId OrgId, loginOrEmail string, role Role) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v/users", c.cfg.Url, orgId))
	c.addAuthorizationHeader(req)

	type requestBody struct {
		LoginOrEmail string `json:"loginOrEmail"`
		Role         string `json:"role"`
	}
	jsonBody, err := json.Marshal(requestBody{
		LoginOrEmail: loginOrEmail,
		Role:         role.String(),
	})
	if err != nil {
		panic(err)
	}

	req.SetBody(jsonBody)
	req.Header.SetContentType("application/json")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to add user to an organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() == 409 && strings.Contains(string(resp.Body()), "User is already member of this organization") {
		return ErrGrafanaUserAlreadyInOrg
	} else if resp.StatusCode() == 404 {
		return ErrGrafanaUserNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add grafana user to organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// ListOrgUsers returns the list of users part of the given organization.
func (c Client) ListOrgUsers(ctx context.Context, orgId OrgId) ([]OrgUser, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v/users", c.cfg.Url, orgId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to query grafana to remove user from an organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to remove grafana user from organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	respBody := []OrgUser{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	// An organization must have an admin user, if list is empty
	// it means that organization doesn't exist.
	if len(respBody) == 0 {
		return nil, ErrGrafanaOrgNotFound
	}

	return respBody, nil
}

// UpdateUserRole updates role of the given user in the given organization.
func (c Client) UpdateUserRole(ctx context.Context, orgId OrgId, userId UserId, role Role) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("PATCH")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v/users/%v", c.cfg.Url, orgId, userId))

	type requestBody struct {
		Role string `json:"role"`
	}

	body, err := json.Marshal(requestBody{
		Role: role.String(),
	})
	if err != nil {
		panic(err)
	}

	req.SetBody(body)
	req.Header.SetContentType("application/json")
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to update user role in organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to update grafana user role in organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// RemoveUserFromOrg removes the given user from the given organization.
func (c Client) RemoveUserFromOrg(ctx context.Context, orgId OrgId, userId UserId) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v/users/%v", c.cfg.Url, orgId, userId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to remove user from an organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to remove grafana user from organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// DeleteOrg deletes organization with the given id.
func (c Client) DeleteOrg(ctx context.Context, orgId OrgId) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/orgs/%v", c.cfg.Url, orgId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to delete organization: %w", err)
	}

	// Grafana returns 403 forbidden when an organization doesn't exist.
	if resp.StatusCode() == 403 && strings.Contains(string(resp.Body()), "You'll need additional permissions to perform this action.") {
		return ErrGrafanaOrgNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// changeCurrentOrg changes current/focused/active organization of client user.
// This is required as some ressources are tied to an organization and doesn't take
// an org id parameter. Grafana deduce organization ID from users current organization.
// This method should be called with mutex locked.
func (c Client) changeCurrentOrg(ctx context.Context, orgId OrgId) error {
	if c.Mutex.TryLock() {
		panic("change current org called without lock")
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/user/using/%v", c.cfg.Url, orgId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to change organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to change grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// CurrentOrg returns current/focused/active organization of client user.
func (c Client) CurrentOrg(ctx context.Context) (OrgId, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/user", c.cfg.Url))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to query grafana to change organization: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to change grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	type responseBody struct {
		OrgId OrgId `json:"orgId"`
	}
	respBody := responseBody{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return 0, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return respBody.OrgId, nil
}

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
	ErrGrafanaDatasourceAlreadyExists  = errors.New("grafana datasource already exists")
	ErrGrafanaDatasourceNotFound       = errors.New("grafana datasource not found")
	ErrGrafanaDatasourceAlreadyUpdated = errors.New("grafana datasource already updated")
)

// Datasource define data sources for grafana dashboards.
type Datasource struct {
	Access         string         `json:"access,omitempty"`
	BasicAuth      bool           `json:"basicAuth,omitempty"`
	Database       string         `json:"database,omitempty"`
	Id             int64          `json:"id,omitempty"`
	IsDefault      bool           `json:"isDefault,omitempty"`
	JSONData       map[string]any `json:"jsonData,omitempty"`
	SecureJSONData map[string]any `json:"secureJsonData"`
	Name           string         `json:"name,omitempty"`
	OrgId          OrgId          `json:"orgId,omitempty"`
	ReadOnly       bool           `json:"readOnly,omitempty"`
	Type           string         `json:"type,omitempty"`
	TypeLogoUrl    string         `json:"typeLogoUrl,omitempty"`
	TypeName       string         `json:"typeName,omitempty"`
	Uid            Uid            `json:"uid,omitempty"`
	URL            string         `json:"url,omitempty"`
	User           string         `json:"user,omitempty"`
	Version        uint           `json:"version,omitempty"`
}

// CreateDatasource creates a datasource in the given organization.
func (c Client) CreateDatasource(ctx context.Context, orgId OrgId, name string, srcType string, isDefault bool) (Datasource, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(c.cfg.Url + "/api/datasources")
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	type requestBody struct {
		Access    string `json:"access"`
		IsDefault bool   `json:"isDefault"`
		Name      string `json:"name"`
		Type      string `json:"type"`
	}
	jsonBody, err := json.Marshal(requestBody{
		Access:    "proxy",
		IsDefault: isDefault,
		Name:      name,
		Type:      srcType,
	})
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
		return Datasource{}, fmt.Errorf("failed to query grafana to create a datasource: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() == 409 && strings.Contains(string(resp.Body()), "data source with the same name already exists") {
		return Datasource{}, ErrGrafanaDatasourceAlreadyExists
	} else if resp.StatusCode() != 200 {
		return Datasource{}, fmt.Errorf("failed to create grafana datasource: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	type responseBody struct {
		Datasource Datasource `json:"datasource"`
	}
	respBody := responseBody{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return Datasource{}, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return respBody.Datasource, nil
}

// UpdateDatasource updates datasource in the given organization.
func (c Client) UpdateDatasource(ctx context.Context, orgId OrgId, datasource Datasource) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("PUT")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/uid/%v", c.cfg.Url, datasource.Uid))
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	jsonBody, err := json.Marshal(datasource)
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
		return fmt.Errorf("failed to query grafana to update a datasource: %w", err)
	}

	// Handle errors.
	switch resp.StatusCode() {
	case 200:
	case 409:
		if strings.Contains(string(resp.Body()), "Datasource has already been updated by someone else. Please reload and try again") {
			return ErrGrafanaDatasourceAlreadyUpdated
		}
		fallthrough
	default:
		return fmt.Errorf("failed to update grafana datasource: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	respBody := Datasource{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return nil
}

// ListDatasources returns a list of datasource present in the given organization.
func (c Client) ListDatasources(ctx context.Context, orgId OrgId) ([]Datasource, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources", c.cfg.Url))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to query grafana to update a datasource: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to create grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	var respBody []Datasource
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return respBody, nil
}

// DeleteDatasourceByName deletes datasource with the given name inside the given organization.
func (c Client) DeleteDatasourceByName(ctx context.Context, orgId OrgId, name string) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/name/%v", c.cfg.Url, name))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to update a datasource: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() == 404 {
		return ErrGrafanaDatasourceNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to create grafana organization: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// GetDatasourceByName retrieves datasource with the given name.
func (c Client) GetDatasourceByName(ctx context.Context, orgId OrgId, name string) (Datasource, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/name/%v", c.cfg.Url, name))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return Datasource{}, fmt.Errorf("failed to query grafana to retrieve a datasource by name: %w", err)
	}

	// Handle errors.
	if resp.StatusCode() == 404 {
		return Datasource{}, ErrGrafanaDatasourceNotFound
	} else if resp.StatusCode() != 200 {
		return Datasource{}, fmt.Errorf("failed to retrieve grafana datasource by name: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	respBody := Datasource{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return Datasource{}, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return respBody, nil
}

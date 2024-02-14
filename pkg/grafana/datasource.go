package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var (
	ErrGrafanaDatasourceAlreadyExists = errors.New("grafana datasource already exists")
	ErrGrafanaDatasourceNotFound      = errors.New("grafana datasource not found")
)

// DatasourceId define a unique dashboard identifier.
type DatasourceId uuid.UUID

// ParseDatasourceId parses the given string and return a DatasourceId if its valid.
// A valid DatasourceId is a valid UUID v4.
func ParseDatasourceId(datasourceID string) (DatasourceId, error) {
	id, err := uuid.Parse(datasourceID)
	if err != nil {
		return DatasourceId{}, err
	}

	return DatasourceId(id), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (did *DatasourceId) UnmarshalJSON(rawJSON []byte) error {
	rawJSON = bytes.TrimPrefix(rawJSON, []byte(`"`))
	rawJSON = bytes.TrimSuffix(rawJSON, []byte(`"`))

	var err error
	*did, err = ParseDatasourceId(string(rawJSON))
	if err != nil {
		return err
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (did DatasourceId) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(did))
}

// String implements fmt.Stringer.
func (did DatasourceId) String() string {
	return uuid.UUID(did).String()
}

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
	Uid            DatasourceId   `json:"uid,omitempty"`
	URL            string         `json:"url,omitempty"`
	User           string         `json:"user,omitempty"`
	Version        uint           `json:"version,omitempty"`
}

// CreateDatasource creates a datasource in the given organization. This method
// rely on user context and therefor, client mutex.
func (c Client) CreateDatasource(ctx context.Context, orgId OrgId, name string, srcType string, isDefault bool) (Datasource, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return Datasource{}, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(c.cfg.Url + "/api/datasources")

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

// UpdateDatasource updates datasource in the given organization. This method
// rely on user context and therefor, client mutex.
func (c Client) UpdateDatasource(ctx context.Context, orgId OrgId, datasource Datasource) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("PUT")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/uid/%v", c.cfg.Url, datasource.Uid))

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
	if resp.StatusCode() != 200 {
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
// This method rely on user context and therefor, client mutex.
func (c Client) ListDatasources(ctx context.Context, orgId OrgId) ([]Datasource, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return nil, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources", c.cfg.Url))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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
// This method rely on user context and therefor, client mutex.
func (c Client) DeleteDatasourceByName(ctx context.Context, orgId OrgId, name string) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/name/%v", c.cfg.Url, name))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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
// This method rely on user context and therefor, client mutex.
func (c Client) GetDatasourceByName(ctx context.Context, orgId OrgId, name string) (Datasource, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return Datasource{}, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/datasources/name/%v", c.cfg.Url, name))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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

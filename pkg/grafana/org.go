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
)

// OrgId define a grafana organization id.
type OrgId int64

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

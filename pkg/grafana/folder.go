package grafana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var (
	ErrGrafanaFolderNotFound = errors.New("folder not found")
)

type Folder struct {
	Id        int64     `json:"id"`
	ParentUid uuid.UUID `json:"parentUid"`
	Title     string    `json:"title"`
	Uid       Uid       `json:"uid"`
}

type FolderPermission struct {
	Permission FolderPermissionLevel `json:"permission"`
	Role       Role                  `json:"role,omitempty"`
	TeamId     int64                 `json:"teamId,omitempty"`
	UserId     UserId                `json:"userId,omitempty"`
}

type SearchFolderResult struct {
	Uid   Uid    `json:"uid"`
	Title string `json:"title"`
	// This is just dashboard path but we match schema of API response.
	Url string `json:"url"`
}

// FolderPermissionLevel enumerate possible folder permission level.
type FolderPermissionLevel int8

const (
	FolderPermissionLevelView FolderPermissionLevel = 1 << iota
	FolderPermissionLevelEdit
	FolderPermissionLevelAdmin
)

// String implements fmt.Stringer.
func (fpl FolderPermissionLevel) String() string {
	switch fpl {
	case FolderPermissionLevelView:
		return "View"
	case FolderPermissionLevelEdit:
		return "Edit"
	case FolderPermissionLevelAdmin:
		return "Admin"
	default:
		panic(fmt.Errorf("unknown folder permission level: %v", int8(fpl)))
	}
}

// CreateFolder creates a folder within current organization.
func (c Client) CreateFolder(ctx context.Context, orgId OrgId, title string) (Folder, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders", c.cfg.Url))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	type requestBody struct {
		Title string `json:"title"`
	}
	jsonBody, err := json.Marshal(requestBody{
		Title: title,
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
		return Folder{}, fmt.Errorf("failed to query grafana to create folder: %w", err)
	}

	if resp.StatusCode() != 200 {
		return Folder{}, fmt.Errorf("failed to create grafana folders: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	var folder Folder
	err = json.Unmarshal(resp.Body(), &folder)
	if err != nil {
		return Folder{}, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return folder, nil
}

// ListFolders lists up to the given limit, children folders of parent folder with
// the given folder UUID.
func (c Client) ListFolders(ctx context.Context, orgId OrgId, limit int, page int) ([]Folder, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/?limit=%v&page=%v", c.cfg.Url, limit, page))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to query grafana to list folders: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list grafana folders: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	var folders []Folder
	err = json.Unmarshal(resp.Body(), &folders)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return folders, nil
}

// GetFolderPermissions gets permissions associated to folder with the given
// Uid.
func (c Client) GetFolderPermissions(ctx context.Context, orgId OrgId, folderId Uid) ([]FolderPermission, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v/permissions", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to query grafana to list folder permissions: %w", err)
	}

	if resp.StatusCode() == 404 {
		return nil, ErrGrafanaFolderNotFound
	} else if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list grafana folder permissions: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	var permissions []FolderPermission
	err = json.Unmarshal(resp.Body(), &permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return permissions, nil
}

// SetFolderPermissions sets permissions associated to folder with the given
// Uid. This operation will remove existing permissions if they're not included
// in the request.
func (c Client) SetFolderPermissions(ctx context.Context, orgId OrgId, folderId Uid, permissions ...FolderPermission) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v/permissions", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	type requestBody struct {
		Items []FolderPermission `json:"items"`
	}
	jsonBody, err := json.Marshal(requestBody{
		Items: permissions,
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
		return fmt.Errorf("failed to query grafana to set folder permissions: %w", err)
	}

	if resp.StatusCode() == 404 {
		return ErrGrafanaFolderNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to set grafana folder permissions: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// DeleteFolder deletes folder with the given Uid.
func (c Client) DeleteFolder(ctx context.Context, orgId OrgId, folderId Uid) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return fmt.Errorf("failed to query grafana to delete folder: %w", err)
	}

	if resp.StatusCode() == 404 {
		return ErrGrafanaFolderNotFound
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to set grafana folder permissions: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	return nil
}

// SearchFolders searches folders within the given organization.
func (c Client) SearchFolders(ctx context.Context, orgId OrgId, limit, page int, query string) ([]SearchFolderResult, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/search?type=dash-folder&limit=%v&page=%v&query=%v", c.cfg.Url, limit, page, url.QueryEscape(query)))
	c.addAuthorizationHeader(req)
	req.Header.Set(GrafanaOrgIdHeader, fmt.Sprint(orgId))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to query grafana to search folders: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to search grafana folders: %v %v", resp.StatusCode(), string(resp.Body()))
	}

	var respBody []SearchFolderResult
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse grafana response: %w", err)
	}

	return respBody, nil
}

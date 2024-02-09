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
	ErrGrafanaFolderAlreadyExists = errors.New("folder with same title already exists")
	ErrGrafanaFolderNotFound      = errors.New("folder not found")
)

// FolderID define a unique dashboard identifier.
type FolderID uuid.UUID

// ParseFolderID parses the given string and return a FolderID if its valid.
// A valid FolderID is a valid UUID v4.
func ParseFolderID(folderID string) (FolderID, error) {
	id, err := uuid.Parse(folderID)
	if err != nil {
		return FolderID{}, err
	}

	return FolderID(id), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (fid *FolderID) UnmarshalJSON(rawJSON []byte) error {
	rawJSON = bytes.TrimPrefix(rawJSON, []byte(`"`))
	rawJSON = bytes.TrimSuffix(rawJSON, []byte(`"`))

	if len(rawJSON) == 0 {
		return nil
	}

	var err error
	*fid, err = ParseFolderID(string(rawJSON))
	if err != nil {
		return err
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (fid FolderID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(fid))
}

// String implements fmt.Stringer.
func (fid FolderID) String() string {
	return uuid.UUID(fid).String()
}

type Folder struct {
	Id        int64     `json:"id"`
	ParentUid uuid.UUID `json:"parentUid"`
	Title     string    `json:"title"`
	Uid       FolderID  `json:"uid"`
}

type FolderPermission struct {
	Permission FolderPermissionLevel `json:"permission"`
	Role       Role                  `json:"role,omitempty"`
	TeamId     int64                 `json:"teamId,omitempty"`
	UserId     UserId                `json:"userId,omitempty"`
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
// This method rely on user context and therefor, client mutex.
func (c Client) CreateFolder(ctx context.Context, orgId OrgId, title string) (Folder, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return Folder{}, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders", c.cfg.Url))
	c.addAuthorizationHeader(req)

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

	if resp.StatusCode() == 409 && strings.Contains(string(resp.Body()), "folder with the same name already exists") {
		return Folder{}, ErrGrafanaFolderAlreadyExists
	} else if resp.StatusCode() != 200 {
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
// This method rely on user context and therefor, client mutex.
func (c Client) ListFolders(ctx context.Context, orgId OrgId, limit int, page int) ([]Folder, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return nil, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/?limit=%v&page=%v", c.cfg.Url, limit, page))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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
// FolderID.
// This method rely on user context and therefor, client mutex.
func (c Client) GetFolderPermissions(ctx context.Context, orgId OrgId, folderId FolderID) ([]FolderPermission, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return nil, fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("GET")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v/permissions", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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
// FolderID. This operation will remove existing permissions if they're not included
// in the request.
// This method rely on user context and therefor, client mutex.
func (c Client) SetFolderPermissions(ctx context.Context, orgId OrgId, folderId FolderID, permissions ...FolderPermission) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v/permissions", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)

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

// DeleteFolder deletes folder with the given FolderID.
// This method rely on user context and therefor, client mutex.
func (c Client) DeleteFolder(ctx context.Context, orgId OrgId, folderId FolderID) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := c.changeCurrentOrg(ctx, orgId)
	if err != nil {
		return fmt.Errorf("failed to change current org: %w", err)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("DELETE")
	req.SetRequestURI(fmt.Sprintf("%v/api/folders/%v", c.cfg.Url, folderId))
	c.addAuthorizationHeader(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.do(ctx, req, resp)
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

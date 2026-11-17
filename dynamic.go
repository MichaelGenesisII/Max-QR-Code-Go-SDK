package getmaxqr

import (
	"context"
	"net/http"
	"net/url"
)

// DynamicService manages dynamic short links.
type DynamicService struct {
	client *Client
}

// List returns dynamic codes for a workspace.
func (s *DynamicService) List(ctx context.Context, workspaceID string, status, q string) (map[string]any, error) {
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	if q != "" {
		query.Set("q", q)
	}
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes"
	if err := s.client.requestJSON(ctx, http.MethodGet, path, query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a dynamic code.
func (s *DynamicService) Create(ctx context.Context, workspaceID string, input CreateDynamicInput) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes"
	if err := s.client.requestJSON(ctx, http.MethodPost, path, nil, input, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get fetches one dynamic code.
func (s *DynamicService) Get(ctx context.Context, workspaceID, codeID string) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes/" + url.PathEscape(codeID)
	if err := s.client.requestJSON(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update patches a dynamic code.
func (s *DynamicService) Update(ctx context.Context, workspaceID, codeID string, input UpdateDynamicInput) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes/" + url.PathEscape(codeID)
	if err := s.client.requestJSON(ctx, http.MethodPatch, path, nil, input, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Deactivate deactivates a dynamic code.
func (s *DynamicService) Deactivate(ctx context.Context, workspaceID, codeID string) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes/" + url.PathEscape(codeID) + "/deactivate"
	if err := s.client.requestJSON(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Activate activates a dynamic code.
func (s *DynamicService) Activate(ctx context.Context, workspaceID, codeID string) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes/" + url.PathEscape(codeID) + "/activate"
	if err := s.client.requestJSON(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a dynamic code.
func (s *DynamicService) Delete(ctx context.Context, workspaceID, codeID string) error {
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/dynamic-codes/" + url.PathEscape(codeID)
	res, err := s.client.raw(ctx, http.MethodDelete, path, nil, nil, "")
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

package getmaxqr

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// BulkService manages bulk CSV/row jobs.
type BulkService struct {
	client *Client
}

// List returns bulk jobs for a workspace.
func (s *BulkService) List(ctx context.Context, workspaceID string) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs"
	if err := s.client.requestJSON(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Create starts a bulk job (requires rows or csv).
func (s *BulkService) Create(ctx context.Context, workspaceID string, input CreateBulkInput) (map[string]any, error) {
	if len(input.Rows) == 0 && input.CSV == "" {
		return nil, &Error{Message: "bulk.Create requires Rows or CSV", Status: 400, Code: "invalid_request"}
	}
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs"
	if err := s.client.requestJSON(ctx, http.MethodPost, path, nil, input, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get fetches one bulk job.
func (s *BulkService) Get(ctx context.Context, workspaceID, jobID string) (map[string]any, error) {
	var out map[string]any
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs/" + url.PathEscape(jobID)
	if err := s.client.requestJSON(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Wait polls until the job completes, fails, or times out.
func (s *BulkService) Wait(ctx context.Context, workspaceID, jobID string, opts WaitOptions) (map[string]any, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	interval := opts.Interval
	if interval == 0 {
		interval = 1500 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)

	for {
		result, err := s.Get(ctx, workspaceID, jobID)
		if err != nil {
			return nil, err
		}
		job, _ := result["job"].(map[string]any)
		status, _ := job["status"].(string)
		switch status {
		case "completed":
			return result, nil
		case "failed", "cancelled":
			msg, _ := job["error_message"].(string)
			if msg == "" {
				msg = "Bulk job " + status
			}
			return nil, &Error{Message: msg, Status: 409, Code: "not_ready", Details: job}
		}
		if time.Now().After(deadline) {
			return nil, &Error{
				Message: "Timed out waiting for bulk job " + jobID,
				Status:  408,
				Code:    "not_ready",
				Details: job,
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// DownloadZip downloads the completed job ZIP.
func (s *BulkService) DownloadZip(ctx context.Context, workspaceID, jobID string) ([]byte, error) {
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs/" + url.PathEscape(jobID) + "/download"
	res, err := s.client.raw(ctx, http.MethodGet, path, nil, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

// Mapping downloads the job mapping CSV/text.
func (s *BulkService) Mapping(ctx context.Context, workspaceID, jobID string) (string, error) {
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs/" + url.PathEscape(jobID) + "/mapping"
	res, err := s.client.raw(ctx, http.MethodGet, path, nil, nil, "")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Cancel cancels a bulk job.
func (s *BulkService) Cancel(ctx context.Context, workspaceID, jobID string) error {
	path := "/workspaces/" + url.PathEscape(workspaceID) + "/bulk-jobs/" + url.PathEscape(jobID)
	res, err := s.client.raw(ctx, http.MethodDelete, path, nil, nil, "")
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

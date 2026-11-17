package getmaxqr

import (
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.getmaxqr.com/api/v1"

// Options configures a Client.
type Options struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// GenerateOptions controls generate / generateWithLogo.
type GenerateOptions struct {
	Format    string         // png | svg | pdf (default png)
	Style     map[string]any // optional style object
	Preset    string
	Response  string // json (default) | binary
	LogoScale *float64
}

// GenerateJSONResult is returned when Response is "json" (default).
type GenerateJSONResult struct {
	Format     string         `json:"format"`
	Mime       string         `json:"mime"`
	Meta       map[string]any `json:"meta,omitempty"`
	Auth       string         `json:"auth,omitempty"`
	DataBase64 string         `json:"data_base64"`
}

// GenerateBinaryResult is returned when Response is "binary".
type GenerateBinaryResult struct {
	Format string
	Mime   string
	Buffer []byte
}

// LogoInput is raw logo bytes for multipart upload.
type LogoInput struct {
	Data     []byte
	Filename string
	Mime     string
}

// WaitOptions controls bulk.Wait polling.
type WaitOptions struct {
	Timeout  time.Duration // default 120s
	Interval time.Duration // default 1.5s
}

// CreateDynamicInput is the body for dynamic.Create.
type CreateDynamicInput struct {
	Name           string  `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	DestinationURL string  `json:"destination_url"`
	Slug           string  `json:"slug,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	CampaignID     string  `json:"campaign_id,omitempty"`
}

// UpdateDynamicInput is the body for dynamic.Update.
type UpdateDynamicInput struct {
	Name           string  `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	DestinationURL string  `json:"destination_url,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	Status         string  `json:"status,omitempty"`
}

// CreateBulkInput is the body for bulk.Create.
type CreateBulkInput struct {
	Name        string           `json:"name,omitempty"`
	ContentType string           `json:"content_type"`
	Format      string           `json:"format,omitempty"`
	Rows        []map[string]any `json:"rows,omitempty"`
	CSV         string           `json:"csv,omitempty"`
	Style       map[string]any   `json:"style,omitempty"`
	Preset      string           `json:"preset,omitempty"`
}

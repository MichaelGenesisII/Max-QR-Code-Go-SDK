package getmaxqr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is the Max QR Code API client.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client

	Dynamic *DynamicService
	Bulk    *BulkService
}

// New creates a Client. APIKey is required for dynamic/bulk (Business mq_live_…).
func New(opts Options) *Client {
	base := strings.TrimRight(opts.BaseURL, "/")
	if base == "" {
		base = DefaultBaseURL
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	c := &Client{
		APIKey:     strings.TrimSpace(opts.APIKey),
		BaseURL:    base,
		HTTPClient: httpClient,
	}
	c.Dynamic = &DynamicService{client: c}
	c.Bulk = &BulkService{client: c}
	return c
}

// GenerateOptionsJSON calls GET /generate/options.
func (c *Client) GenerateOptions(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.requestJSON(ctx, http.MethodGet, "/generate/options", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Generate calls POST /generate. Default response is JSON with data_base64.
// Pass opts.Response == "binary" for raw image bytes.
func (c *Client) Generate(ctx context.Context, content Content, opts GenerateOptions) (any, error) {
	format := opts.Format
	if format == "" {
		format = "png"
	}
	responseMode := opts.Response
	if responseMode == "" {
		responseMode = "json"
	}

	body := map[string]any{
		"content":  content,
		"format":   format,
		"response": responseMode,
	}
	if opts.Style != nil {
		body["style"] = opts.Style
	}
	if opts.Preset != "" {
		body["preset"] = opts.Preset
	}

	if responseMode == "binary" {
		delete(body, "response")
		res, err := c.raw(ctx, http.MethodPost, "/generate", body, nil, "application/json")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		mime := res.Header.Get("Content-Type")
		if mime == "" {
			mime = "application/octet-stream"
		}
		return GenerateBinaryResult{
			Format: guessFormat(mime, format),
			Mime:   mime,
			Buffer: buf,
		}, nil
	}

	var out GenerateJSONResult
	q := url.Values{"response": {"json"}}
	if err := c.requestJSON(ctx, http.MethodPost, "/generate", q, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GenerateWithLogo calls POST /generate/upload (multipart). Prefer the API host for large logos.
func (c *Client) GenerateWithLogo(ctx context.Context, content Content, logo LogoInput, opts GenerateOptions) (any, error) {
	format := opts.Format
	if format == "" {
		format = "png"
	}
	responseMode := opts.Response
	if responseMode == "" {
		responseMode = "json"
	}
	filename := logo.Filename
	if filename == "" {
		filename = "logo.png"
	}
	mime := logo.Mime
	if mime == "" {
		mime = "application/octet-stream"
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("content", mustJSON(content))
	_ = w.WriteField("format", format)
	_ = w.WriteField("response", responseMode)
	if opts.Style != nil {
		_ = w.WriteField("style", mustJSON(opts.Style))
	}
	if opts.Preset != "" {
		_ = w.WriteField("preset", opts.Preset)
	}
	if opts.LogoScale != nil {
		_ = w.WriteField("logo_scale", strconv.FormatFloat(*opts.LogoScale, 'f', -1, 64))
	}
	part, err := w.CreateFormFile("logo", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(logo.Data); err != nil {
		return nil, err
	}
	_ = w.Close()

	path := "/generate/upload"
	var q url.Values
	if responseMode == "json" {
		q = url.Values{"response": {"json"}}
	}

	res, err := c.rawMultipart(ctx, http.MethodPost, path, q, &buf, w.FormDataContentType())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if responseMode == "binary" {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		outMime := res.Header.Get("Content-Type")
		if outMime == "" {
			outMime = "application/octet-stream"
		}
		return GenerateBinaryResult{
			Format: guessFormat(outMime, format),
			Mime:   outMime,
			Buffer: data,
		}, nil
	}

	var out GenerateJSONResult
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, &Error{Message: "API returned non-JSON response", Status: res.StatusCode, Code: "api_error"}
	}
	return out, nil
}

func (c *Client) requestJSON(ctx context.Context, method, path string, query url.Values, body any, dest any) error {
	res, err := c.raw(ctx, method, path, body, query, "application/json")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return &Error{
			Message:   "API returned non-JSON response",
			Status:    res.StatusCode,
			Code:      "api_error",
			RequestID: res.Header.Get("X-Request-Id"),
		}
	}
	return nil
}

func (c *Client) raw(ctx context.Context, method, path string, body any, query url.Values, contentType string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	return c.do(ctx, method, path, query, bodyReader, contentType)
}

func (c *Client) rawMultipart(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string) (*http.Response, error) {
	return c.do(ctx, method, path, query, body, contentType)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType string) (*http.Response, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.APIKey != "" {
		req.Header.Set("X-Api-Key", c.APIKey)
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return nil, throwFromResponse(res)
	}
	return res, nil
}

func throwFromResponse(res *http.Response) error {
	retryAfter := parseRetryAfter(res.Header.Get("Retry-After"))
	requestID := res.Header.Get("X-Request-Id")
	message := fmt.Sprintf("Max QR API error (%d)", res.StatusCode)
	code := "api_error"
	var details any

	data, _ := io.ReadAll(res.Body)
	if len(data) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err == nil {
			if errObj, ok := payload["error"].(map[string]any); ok {
				if m, ok := errObj["message"].(string); ok && m != "" {
					message = m
				}
				if c, ok := errObj["code"].(string); ok && c != "" {
					code = c
				}
				if d, ok := errObj["details"]; ok {
					details = d
				} else {
					details = payload
				}
			} else if m, ok := payload["message"].(string); ok && m != "" {
				message = m
				details = payload
			} else {
				details = payload
			}
		} else {
			message = string(data)
			if len(message) > 500 {
				message = message[:500]
			}
		}
	}

	return &Error{
		Message:           message,
		Status:            res.StatusCode,
		Code:              code,
		Details:           details,
		RetryAfterSeconds: retryAfter,
		RequestID:         requestID,
	}
}

func parseRetryAfter(v string) *float64 {
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func guessFormat(mime, fallback string) string {
	if strings.Contains(mime, "svg") {
		return "svg"
	}
	if strings.Contains(mime, "pdf") {
		return "pdf"
	}
	return fallback
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

package getmaxqr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateSendsAPIKeyAndParsesJSON(t *testing.T) {
	var gotKey, gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		gotURL = r.URL.String()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"format":      "png",
			"mime":        "image/png",
			"data_base64": "abc",
			"auth":        "api_key",
		})
	}))
	defer srv.Close()

	c := New(Options{APIKey: "mq_live_test", BaseURL: srv.URL})
	out, err := c.Generate(context.Background(), URL("https://example.com"), GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	res := out.(GenerateJSONResult)
	if res.DataBase64 != "abc" {
		t.Fatalf("data_base64=%q", res.DataBase64)
	}
	if gotKey != "mq_live_test" {
		t.Fatalf("api key=%q", gotKey)
	}
	if !strings.Contains(gotURL, "/generate?response=json") {
		t.Fatalf("url=%q", gotURL)
	}
}

func TestRateLimitedWithRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "12")
		w.Header().Set("X-Request-Id", "req_1")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "rate_limited", "message": "Slow down"},
		})
	}))
	defer srv.Close()

	c := New(Options{APIKey: "mq_live_test", BaseURL: srv.URL})
	_, err := c.GenerateOptions(context.Background())
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T %v", err, err)
	}
	if apiErr.Code != "rate_limited" || !apiErr.IsRateLimited() {
		t.Fatalf("%+v", apiErr)
	}
	if apiErr.RetryAfterSeconds == nil || *apiErr.RetryAfterSeconds != 12 {
		t.Fatalf("retry=%v", apiErr.RetryAfterSeconds)
	}
	if apiErr.RequestID != "req_1" {
		t.Fatalf("request id=%q", apiErr.RequestID)
	}
}

func TestDynamicCreatePostsDestinationURL(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": map[string]any{
				"id":              "c1",
				"redirect_url":    "https://go.getmaxqr.com/r/menu",
				"destination_url": "https://example.com/menu",
			},
		})
	}))
	defer srv.Close()

	c := New(Options{APIKey: "mq_live_test", BaseURL: srv.URL})
	out, err := c.Dynamic.Create(context.Background(), "w1", CreateDynamicInput{
		Name:           "Menu",
		DestinationURL: "https://example.com/menu",
	})
	if err != nil {
		t.Fatal(err)
	}
	code := out["code"].(map[string]any)
	if code["redirect_url"] != "https://go.getmaxqr.com/r/menu" {
		t.Fatalf("%+v", code)
	}
	if !strings.Contains(body, "destination_url") {
		t.Fatalf("body=%s", body)
	}
}

func TestBulkWaitResolvesWhenCompleted(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		status := "running"
		hasZip := false
		if n > 1 {
			status = "completed"
			hasZip = true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"job": map[string]any{
				"id":      "j1",
				"status":  status,
				"has_zip": hasZip,
			},
		})
	}))
	defer srv.Close()

	c := New(Options{APIKey: "mq_live_test", BaseURL: srv.URL})
	out, err := c.Bulk.Wait(context.Background(), "w1", "j1", WaitOptions{
		Timeout:  time.Second,
		Interval: time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	job := out["job"].(map[string]any)
	if job["status"] != "completed" || job["has_zip"] != true {
		t.Fatalf("%+v", job)
	}
}

func TestGenerateWithLogoSendsMultipart(t *testing.T) {
	var contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"format":      "png",
			"mime":        "image/png",
			"data_base64": "xyz",
		})
	}))
	defer srv.Close()

	c := New(Options{APIKey: "mq_live_test", BaseURL: srv.URL})
	out, err := c.GenerateWithLogo(context.Background(), URL("https://example.com"), LogoInput{
		Data:     []byte{1, 2, 3},
		Filename: "mark.png",
		Mime:     "image/png",
	}, GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	res := out.(GenerateJSONResult)
	if res.DataBase64 != "xyz" {
		t.Fatalf("%+v", res)
	}
	if !strings.Contains(contentType, "multipart/form-data") {
		t.Fatalf("content-type=%q", contentType)
	}
}

func TestBulkCreateRequiresRowsOrCSV(t *testing.T) {
	c := New(Options{APIKey: "mq_live_test", BaseURL: "http://example.invalid"})
	_, err := c.Bulk.Create(context.Background(), "w1", CreateBulkInput{ContentType: "url"})
	apiErr, ok := err.(*Error)
	if !ok || apiErr.Code != "invalid_request" {
		t.Fatalf("%v", err)
	}
}

package getmaxqr

import (
	"fmt"
	"net/http"
)

// Error is an API or client error from the Max QR Code SDK.
type Error struct {
	Message           string
	Status            int
	Code              string
	Details           any
	RetryAfterSeconds *float64
	RequestID         string
}

func (e *Error) Error() string {
	if e == nil {
		return "getmaxqr: unknown error"
	}
	if e.Code != "" {
		return fmt.Sprintf("getmaxqr: %s (%d): %s", e.Code, e.Status, e.Message)
	}
	return fmt.Sprintf("getmaxqr: (%d): %s", e.Status, e.Message)
}

func (e *Error) IsRateLimited() bool {
	return e != nil && (e.Code == "rate_limited" || e.Status == http.StatusTooManyRequests)
}

func (e *Error) IsDailyQuota() bool {
	return e != nil && e.Code == "daily_quota"
}

func (e *Error) IsPlanError() bool {
	return e != nil && (e.Code == "plan_forbidden" || e.Code == "plan_limit")
}

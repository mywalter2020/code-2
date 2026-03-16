package adapters

import "time"

type AdapterTrace struct {
	Platform    string         `json:"platform"`
	Action      string         `json:"action"`
	Mode        string         `json:"mode,omitempty"`
	RequestID   string         `json:"request_id,omitempty"`
	StartedAt   string         `json:"started_at,omitempty"`
	FinishedAt  string         `json:"finished_at,omitempty"`
	DurationMs  int64          `json:"duration_ms,omitempty"`
	Attempts    int            `json:"attempts,omitempty"`
	HTTPStatus  int            `json:"http_status,omitempty"`
	URL         string         `json:"url,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
	Headers     map[string]any `json:"headers,omitempty"`
	Body        map[string]any `json:"body,omitempty"`
	Response    map[string]any `json:"response,omitempty"`
	Error       string         `json:"error,omitempty"`
	ErrorClass  string         `json:"error_class,omitempty"`
}

func newTrace(platform, action, mode, requestID string) AdapterTrace {
	return AdapterTrace{
		Platform:  platform,
		Action:    action,
		Mode:      mode,
		RequestID: requestID,
		StartedAt: time.Now().Format(time.RFC3339Nano),
	}
}

func finishTrace(t AdapterTrace, httpStatus int, response map[string]any, errText, errClass string) AdapterTrace {
	finished := time.Now()
	started, _ := time.Parse(time.RFC3339Nano, t.StartedAt)
	t.FinishedAt = finished.Format(time.RFC3339Nano)
	if !started.IsZero() {
		t.DurationMs = finished.Sub(started).Milliseconds()
	}
	t.HTTPStatus = httpStatus
	t.Response = response
	t.Error = errText
	t.ErrorClass = errClass
	return t
}

package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AdapterHTTPClient struct {
	client HTTPClient
}

type AdapterHTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    any
	Timeout time.Duration
}

type AdapterHTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	JSON       map[string]any
}

func NewAdapterHTTPClient(client HTTPClient) *AdapterHTTPClient {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &AdapterHTTPClient{client: client}
}

func (c *AdapterHTTPClient) DoJSON(ctx context.Context, req AdapterHTTPRequest) (AdapterHTTPResponse, error) {
	method := strings.TrimSpace(strings.ToUpper(req.Method))
	if method == "" {
		method = http.MethodPost
	}
	var bodyReader io.Reader
	if req.Body != nil {
		payload, err := json.Marshal(req.Body)
		if err != nil {
			return AdapterHTTPResponse{}, err
		}
		bodyReader = bytes.NewReader(payload)
	}
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, req.URL, bodyReader)
	if err != nil {
		return AdapterHTTPResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	for k, v := range req.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return AdapterHTTPResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AdapterHTTPResponse{}, err
	}
	out := AdapterHTTPResponse{StatusCode: resp.StatusCode, Body: body, Headers: map[string]string{}}
	for k := range resp.Header {
		out.Headers[k] = resp.Header.Get(k)
	}
	if len(bytes.TrimSpace(body)) > 0 {
		_ = json.Unmarshal(body, &out.JSON)
	}
	if resp.StatusCode >= 300 {
		return out, fmt.Errorf("adapter http status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return out, nil
}

package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"juyu-ai-platform/internal/types"
)

type BaseAdapter struct {
	name            string
	displayName     string
	requiredFields  []string
	credentialStore *CredentialStore
	dryRun          bool
	baseURL         string
	httpClient      *AdapterHTTPClient
}

func NewBaseAdapter(name, displayName string, requiredFields []string, credentialStore *CredentialStore, dryRun bool, baseURL string) BaseAdapter {
	return BaseAdapter{
		name:            name,
		displayName:     displayName,
		requiredFields:  requiredFields,
		credentialStore: credentialStore,
		dryRun:          dryRun,
		baseURL:         strings.TrimSpace(baseURL),
		httpClient:      NewAdapterHTTPClient(nil),
	}
}

func (a BaseAdapter) Name() string { return a.name }

func (a BaseAdapter) Descriptor() types.AdapterDescriptor {
	configured, _ := a.checkCredentials()
	baseURLConfigured := a.baseURL != ""
	return types.AdapterDescriptor{
		Platform:          a.name,
		DisplayName:       a.displayName,
		RequiredFields:    append([]string{}, a.requiredFields...),
		DryRun:            a.dryRun,
		Configured:        configured,
		BaseURLConfigured: baseURLConfigured,
		LiveReady:         configured && baseURLConfigured,
		SupportedActions:  []string{"publish", "update", "on_shelf", "off_shelf"},
	}
}

func (a BaseAdapter) Health(ctx context.Context) (types.AdapterHealth, error) {
	_ = ctx
	configured, missing := a.checkCredentials()
	baseURLConfigured := a.baseURL != ""
	message := "adapter ready"
	if a.dryRun {
		message = "dry-run mode"
	} else if !baseURLConfigured {
		message = "live mode requires base_url"
	}
	if !configured {
		message = fmt.Sprintf("missing credentials: %s", strings.Join(missing, ", "))
	}
	return types.AdapterHealth{
		Platform:          a.name,
		Healthy:           a.dryRun || (configured && baseURLConfigured),
		Configured:        configured,
		DryRun:            a.dryRun,
		BaseURLConfigured: baseURLConfigured,
		LiveReady:         configured && baseURLConfigured,
		Message:           message,
		MissingFields:     missing,
	}, nil
}

func (a BaseAdapter) Respond(action, status string, req types.AdapterRequest) (types.AdapterResponse, error) {
	configured, missing := a.checkCredentials()
	if err := a.validateRequest(req); err != nil {
		return types.AdapterResponse{}, err
	}
	if !configured && !a.dryRun {
		return types.AdapterResponse{}, fmt.Errorf("adapter %s missing credentials: %s", a.name, strings.Join(missing, ", "))
	}
	if !a.dryRun && a.baseURL == "" {
		return types.AdapterResponse{}, fmt.Errorf("adapter %s missing base_url in live mode", a.name)
	}
	mode := "live"
	if a.dryRun {
		mode = "dry_run"
	}
	return types.AdapterResponse{
		Platform:   a.name,
		Action:     action,
		Status:     status,
		Mode:       mode,
		Configured: configured,
		RequestID:  req.RequestID,
		Data: map[string]any{
			"payload":      req.Payload,
			"operator":     req.Operator,
			"configured":   configured,
			"dry_run":      a.dryRun,
			"missing":      missing,
			"base_url":     a.baseURL,
			"external_ref": req.ExternalRef,
		},
	}, nil
}

func (a BaseAdapter) isConfigured() bool {
	ok, _ := a.checkCredentials()
	return ok
}

func (a BaseAdapter) withHTTPClient(client HTTPClient) BaseAdapter {
	a.httpClient = NewAdapterHTTPClient(client)
	return a
}

func (a BaseAdapter) endpoint(path string) string {
	base := strings.TrimRight(a.baseURL, "/")
	if base == "" {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}

func (a BaseAdapter) liveCall(ctx context.Context, action string, req types.AdapterRequest, path string, body any) (types.AdapterResponse, error) {
	if a.httpClient == nil {
		a.httpClient = NewAdapterHTTPClient(nil)
	}
	resp, err := a.httpClient.DoJSON(ctx, AdapterHTTPRequest{
		Method: httpMethodForAction(action),
		URL:    a.endpoint(path),
		Headers: map[string]string{
			"X-Request-ID":   req.RequestID,
			"X-External-Ref": req.ExternalRef,
		},
		Body:    body,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return types.AdapterResponse{}, err
	}
	status := strings.TrimSpace(jsonString(resp.JSON, "status", "success"))
	if status == "" {
		status = "success"
	}
	return types.AdapterResponse{
		Platform:   a.name,
		Action:     action,
		Status:     status,
		Mode:       "live",
		Configured: a.isConfigured(),
		RequestID:  req.RequestID,
		Data: map[string]any{
			"http_status":    resp.StatusCode,
			"response":       resp.JSON,
			"base_url":       a.baseURL,
			"external_ref":   req.ExternalRef,
			"request_action": action,
		},
	}, nil
}

func jsonString(m map[string]any, key, fallback string) string {
	if m == nil {
		return fallback
	}
	if v, ok := m[key].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func (a BaseAdapter) validateRequest(req types.AdapterRequest) error {
	if strings.TrimSpace(req.Platform) == "" {
		return fmt.Errorf("adapter %s request missing platform", a.name)
	}
	if strings.TrimSpace(req.Action) == "" {
		return fmt.Errorf("adapter %s request missing action", a.name)
	}
	if req.Payload == nil {
		return fmt.Errorf("adapter %s request missing payload", a.name)
	}
	product, ok := req.Payload["product"].(types.Product)
	if !ok {
		if productMap, mapOK := req.Payload["product"].(map[string]any); mapOK {
			if title, ok := productMap["title"].(string); !ok || strings.TrimSpace(title) == "" {
				return fmt.Errorf("adapter %s request missing product.title", a.name)
			}
			return nil
		}
		return fmt.Errorf("adapter %s request missing product", a.name)
	}
	if strings.TrimSpace(product.Title) == "" {
		return fmt.Errorf("adapter %s request missing product.title", a.name)
	}
	return nil
}

func (a BaseAdapter) checkCredentials() (bool, []string) {
	if a.credentialStore == nil || len(a.requiredFields) == 0 {
		return true, nil
	}
	creds, ok := a.credentialStore.Get(a.name)
	if !ok {
		return false, append([]string{}, a.requiredFields...)
	}
	missing := make([]string, 0)
	for _, field := range a.requiredFields {
		if strings.TrimSpace(creds.Fields[field]) == "" {
			missing = append(missing, field)
		}
	}
	return len(missing) == 0, missing
}

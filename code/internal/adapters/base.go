package adapters

import (
	"context"
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

type BaseAdapter struct {
	name            string
	displayName     string
	requiredFields  []string
	credentialStore *CredentialStore
	dryRun          bool
}

func NewBaseAdapter(name, displayName string, requiredFields []string, credentialStore *CredentialStore, dryRun bool) BaseAdapter {
	return BaseAdapter{
		name:            name,
		displayName:     displayName,
		requiredFields:  requiredFields,
		credentialStore: credentialStore,
		dryRun:          dryRun,
	}
}

func (a BaseAdapter) Name() string { return a.name }

func (a BaseAdapter) Descriptor() types.AdapterDescriptor {
	return types.AdapterDescriptor{
		Platform:         a.name,
		DisplayName:      a.displayName,
		RequiredFields:   append([]string{}, a.requiredFields...),
		DryRun:           a.dryRun,
		Configured:       a.isConfigured(),
		SupportedActions: []string{"publish", "update", "on_shelf", "off_shelf"},
	}
}

func (a BaseAdapter) Health(ctx context.Context) (types.AdapterHealth, error) {
	_ = ctx
	configured, missing := a.checkCredentials()
	message := "adapter ready"
	if a.dryRun {
		message = "dry-run mode"
	}
	if !configured {
		message = fmt.Sprintf("missing credentials: %s", strings.Join(missing, ", "))
	}
	return types.AdapterHealth{
		Platform:      a.name,
		Healthy:       true,
		Configured:    configured,
		DryRun:        a.dryRun,
		Message:       message,
		MissingFields: missing,
	}, nil
}

func (a BaseAdapter) Respond(action, status string, req types.AdapterRequest) (types.AdapterResponse, error) {
	configured, missing := a.checkCredentials()
	if !configured && !a.dryRun {
		return types.AdapterResponse{}, fmt.Errorf("adapter %s missing credentials: %s", a.name, strings.Join(missing, ", "))
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
		Data: map[string]any{
			"payload":    req.Payload,
			"operator":   req.Operator,
			"configured": configured,
			"dry_run":    a.dryRun,
			"missing":    missing,
		},
	}, nil
}

func (a BaseAdapter) isConfigured() bool {
	ok, _ := a.checkCredentials()
	return ok
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

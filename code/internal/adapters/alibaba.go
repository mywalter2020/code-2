package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type AlibabaAdapter struct{ BaseAdapter }

func NewAlibabaAdapter(store *CredentialStore, dryRun bool) *AlibabaAdapter {
	cfg := LoadPlatformConfigsFromEnv()["alibaba"]
	return &AlibabaAdapter{BaseAdapter: NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, store, dryRun, cfg.BaseURL)}
}

func (a *AlibabaAdapter) Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("publish", "published", req)
	}
	return a.liveInvoke(ctx, "publish", req)
}
func (a *AlibabaAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("update", "updated", req)
	}
	return a.liveInvoke(ctx, "update", req)
}
func (a *AlibabaAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("on_shelf", "on_shelf", req)
	}
	return a.liveInvoke(ctx, "on_shelf", req)
}
func (a *AlibabaAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("off_shelf", "off_shelf", req)
	}
	return a.liveInvoke(ctx, "off_shelf", req)
}

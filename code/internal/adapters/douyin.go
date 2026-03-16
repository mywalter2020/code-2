package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type DouyinAdapter struct{ BaseAdapter }

func NewDouyinAdapter(store *CredentialStore, dryRun bool) *DouyinAdapter {
	cfg := LoadPlatformConfigsFromEnv()["douyin"]
	return &DouyinAdapter{BaseAdapter: NewBaseAdapter("douyin", "Douyin", []string{"client_id", "client_secret"}, store, dryRun, cfg.BaseURL)}
}

func (a *DouyinAdapter) Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("publish", "published", req)
}
func (a *DouyinAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("update", "updated", req)
}
func (a *DouyinAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("on_shelf", "on_shelf", req)
}
func (a *DouyinAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("off_shelf", "off_shelf", req)
}

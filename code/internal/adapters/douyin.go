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
	if a.dryRun {
		return a.Respond("publish", "published", req)
	}
	return a.liveCall(ctx, "publish", req, "/publish", requestPayload(req))
}
func (a *DouyinAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("update", "updated", req)
	}
	return a.liveCall(ctx, "update", req, "/update", requestPayload(req))
}
func (a *DouyinAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("on_shelf", "on_shelf", req)
	}
	return a.liveCall(ctx, "on_shelf", req, "/on_shelf", requestPayload(req))
}
func (a *DouyinAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("off_shelf", "off_shelf", req)
	}
	return a.liveCall(ctx, "off_shelf", req, "/off_shelf", requestPayload(req))
}

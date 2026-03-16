package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type TaobaoAdapter struct{ BaseAdapter }

func NewTaobaoAdapter(store *CredentialStore, dryRun bool) *TaobaoAdapter {
	cfg := LoadPlatformConfigsFromEnv()["taobao"]
	return &TaobaoAdapter{BaseAdapter: NewBaseAdapter("taobao", "Taobao", []string{"app_key", "secret"}, store, dryRun, cfg.BaseURL)}
}

func (a *TaobaoAdapter) Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("publish", "published", req)
	}
	return a.liveCall(ctx, "publish", req, "/publish", requestPayload(req))
}
func (a *TaobaoAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("update", "updated", req)
	}
	return a.liveCall(ctx, "update", req, "/update", requestPayload(req))
}
func (a *TaobaoAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("on_shelf", "on_shelf", req)
	}
	return a.liveCall(ctx, "on_shelf", req, "/on_shelf", requestPayload(req))
}
func (a *TaobaoAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	if a.dryRun {
		return a.Respond("off_shelf", "off_shelf", req)
	}
	return a.liveCall(ctx, "off_shelf", req, "/off_shelf", requestPayload(req))
}

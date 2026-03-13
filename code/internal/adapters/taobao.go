package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type TaobaoAdapter struct{ BaseAdapter }

func NewTaobaoAdapter(store *CredentialStore, dryRun bool) *TaobaoAdapter {
	return &TaobaoAdapter{BaseAdapter: NewBaseAdapter("taobao", "Taobao", []string{"app_key", "secret"}, store, dryRun)}
}

func (a *TaobaoAdapter) Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("publish", "published", req)
}
func (a *TaobaoAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("update", "updated", req)
}
func (a *TaobaoAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("on_shelf", "on_shelf", req)
}
func (a *TaobaoAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	_ = ctx
	return a.Respond("off_shelf", "off_shelf", req)
}

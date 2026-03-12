package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type TaobaoAdapter struct{}

func NewTaobaoAdapter() *TaobaoAdapter { return &TaobaoAdapter{} }
func (a *TaobaoAdapter) Name() string  { return "taobao" }
func (a *TaobaoAdapter) Health(ctx context.Context) (types.AdapterHealth, error) {
	return types.AdapterHealth{Platform: a.Name(), Healthy: true, Message: "stub adapter ready"}, nil
}
func (a *TaobaoAdapter) Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	return types.AdapterResponse{Platform: a.Name(), Action: "publish", Status: "published", Data: map[string]any{"payload": req.Payload, "operator": req.Operator}}, nil
}
func (a *TaobaoAdapter) Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	return types.AdapterResponse{Platform: a.Name(), Action: "update", Status: "updated", Data: map[string]any{"payload": req.Payload, "operator": req.Operator}}, nil
}
func (a *TaobaoAdapter) OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	return types.AdapterResponse{Platform: a.Name(), Action: "on_shelf", Status: "on_shelf", Data: map[string]any{"payload": req.Payload, "operator": req.Operator}}, nil
}
func (a *TaobaoAdapter) OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error) {
	return types.AdapterResponse{Platform: a.Name(), Action: "off_shelf", Status: "off_shelf", Data: map[string]any{"payload": req.Payload, "operator": req.Operator}}, nil
}

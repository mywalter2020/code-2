package adapters

import "context"

type TaobaoAdapter struct{}

func NewTaobaoAdapter() *TaobaoAdapter { return &TaobaoAdapter{} }
func (a *TaobaoAdapter) Name() string  { return "taobao" }

func (a *TaobaoAdapter) Publish(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "publish", "status": "published", "payload": payload}, nil
}

func (a *TaobaoAdapter) Update(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "update", "status": "updated", "payload": payload}, nil
}

func (a *TaobaoAdapter) OnShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "on_shelf", "status": "on_shelf", "payload": payload}, nil
}

func (a *TaobaoAdapter) OffShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "off_shelf", "status": "off_shelf", "payload": payload}, nil
}

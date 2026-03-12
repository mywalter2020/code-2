package adapters

import "context"

type AlibabaAdapter struct{}

func NewAlibabaAdapter() *AlibabaAdapter { return &AlibabaAdapter{} }
func (a *AlibabaAdapter) Name() string   { return "alibaba" }

func (a *AlibabaAdapter) Publish(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "publish", "status": "published", "payload": payload}, nil
}

func (a *AlibabaAdapter) Update(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "update", "status": "updated", "payload": payload}, nil
}

func (a *AlibabaAdapter) OnShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "on_shelf", "status": "on_shelf", "payload": payload}, nil
}

func (a *AlibabaAdapter) OffShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "off_shelf", "status": "off_shelf", "payload": payload}, nil
}

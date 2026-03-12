package adapters

import "context"

type DouyinAdapter struct{}

func NewDouyinAdapter() *DouyinAdapter { return &DouyinAdapter{} }
func (a *DouyinAdapter) Name() string  { return "douyin" }

func (a *DouyinAdapter) Publish(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "publish", "status": "published", "payload": payload}, nil
}

func (a *DouyinAdapter) Update(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "update", "status": "updated", "payload": payload}, nil
}

func (a *DouyinAdapter) OnShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "on_shelf", "status": "on_shelf", "payload": payload}, nil
}

func (a *DouyinAdapter) OffShelf(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return map[string]any{"platform": a.Name(), "action": "off_shelf", "status": "off_shelf", "payload": payload}, nil
}

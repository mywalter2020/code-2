package adapters

import "context"

type PlatformAdapter interface {
	Name() string
	Publish(ctx context.Context, payload map[string]any) (map[string]any, error)
	Update(ctx context.Context, payload map[string]any) (map[string]any, error)
	OnShelf(ctx context.Context, payload map[string]any) (map[string]any, error)
	OffShelf(ctx context.Context, payload map[string]any) (map[string]any, error)
}

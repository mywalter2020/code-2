package adapters

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type PlatformAdapter interface {
	Name() string
	Descriptor() types.AdapterDescriptor
	Health(ctx context.Context) (types.AdapterHealth, error)
	Publish(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error)
	Update(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error)
	OnShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error)
	OffShelf(ctx context.Context, req types.AdapterRequest) (types.AdapterResponse, error)
}

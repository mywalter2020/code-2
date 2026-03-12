package adapters

import "fmt"

type Registry struct {
	items map[string]PlatformAdapter
}

func NewRegistry() *Registry {
	return &Registry{items: map[string]PlatformAdapter{}}
}

func (r *Registry) Register(adapter PlatformAdapter) {
	r.items[adapter.Name()] = adapter
}

func (r *Registry) Get(name string) (PlatformAdapter, error) {
	adapter, ok := r.items[name]
	if !ok {
		return nil, fmt.Errorf("platform adapter not found: %s", name)
	}
	return adapter, nil
}

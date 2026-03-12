package server

import (
	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/types"
)

type adapterRuntime struct{ registry *adapters.Registry }
func (a *adapterRuntime) ListNames() []string { return a.registry.ListNames() }
func (a *adapterRuntime) Get(name string) (interface{}, error) { return a.registry.Get(name) }

type credentialRuntime struct{ store *adapters.CredentialStore }
func (c *credentialRuntime) List() []types.AdapterCredentials { return c.store.List() }

func attachAdapterRuntime(server *Server, registry *adapters.Registry, credentials *adapters.CredentialStore) {
	server.adapterRegistry = &adapterRuntime{registry: registry}
	server.credentialStore = &credentialRuntime{store: credentials}
}

package server

import "juyu-ai-platform/internal/adapters"

func AttachAdapterRuntime(server *Server, registry *adapters.Registry, credentials *adapters.CredentialStore) {
	attachAdapterRuntime(server, registry, credentials)
}

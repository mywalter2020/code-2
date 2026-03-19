package server

import "juyu-ai-platform/internal/store"

func AttachRuntimeStore(server *Server, runtimeStore store.RuntimeStore, driver string) {
	if runtimeStore == nil {
		return
	}
	server.runtimeStore = runtimeStore
	if driver != "" {
		server.runtimeDriver = driver
	}
}

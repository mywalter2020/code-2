package adapters

import (
	"context"
	"testing"

	"juyu-ai-platform/internal/types"
)

func TestBaseAdapterDryRunAllowsMissingCredentialsButValidatesPayload(t *testing.T) {
	a := NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, NewCredentialStore(), true, "")
	_, err := a.Respond("publish", "published", types.AdapterRequest{})
	if err == nil {
		t.Fatal("expected request validation error")
	}
}

func TestBaseAdapterLiveModeRequiresBaseURLAndCredentials(t *testing.T) {
	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{"app_key": "k", "secret": "s"}})
	a := NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, store, false, "")
	_, err := a.Respond("publish", "published", types.AdapterRequest{Platform: "alibaba", Action: "publish", Payload: map[string]any{"product": types.Product{Title: "demo"}}})
	if err == nil {
		t.Fatal("expected missing base_url error")
	}
}

func TestBaseAdapterDescriptorAndHealthExposeLiveReadiness(t *testing.T) {
	store := NewCredentialStore()
	store.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{"app_key": "k", "secret": "s"}})
	a := NewBaseAdapter("alibaba", "Alibaba", []string{"app_key", "secret"}, store, false, "https://api.example.com")
	d := a.Descriptor()
	if !d.LiveReady || !d.BaseURLConfigured {
		t.Fatalf("expected live-ready descriptor, got %+v", d)
	}
	h, err := a.Health(context.Background())
	if err != nil {
		t.Fatalf("health error: %v", err)
	}
	if !h.Healthy || !h.LiveReady {
		t.Fatalf("expected healthy live adapter, got %+v", h)
	}
}

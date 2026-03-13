package adapters

func NewDefaultRegistry(store *CredentialStore, dryRun bool) *Registry {
	registry := NewRegistry()
	registry.Register(NewAlibabaAdapter(store, dryRun))
	registry.Register(NewTaobaoAdapter(store, dryRun))
	registry.Register(NewDouyinAdapter(store, dryRun))
	return registry
}

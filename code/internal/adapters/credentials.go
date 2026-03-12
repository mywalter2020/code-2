package adapters

import "juyu-ai-platform/internal/types"

type CredentialStore struct {
	items map[string]types.AdapterCredentials
}

func NewCredentialStore() *CredentialStore {
	return &CredentialStore{items: map[string]types.AdapterCredentials{}}
}

func (s *CredentialStore) Set(creds types.AdapterCredentials) {
	s.items[creds.Platform] = creds
}

func (s *CredentialStore) Get(platform string) (types.AdapterCredentials, bool) {
	v, ok := s.items[platform]
	return v, ok
}

func (s *CredentialStore) List() []types.AdapterCredentials {
	items := make([]types.AdapterCredentials, 0, len(s.items))
	for _, v := range s.items {
		items = append(items, v)
	}
	return items
}

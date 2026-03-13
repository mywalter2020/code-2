package llm

import "context"

type ContentGenerator interface {
	ProviderName() string
	Enabled() bool
	Config() ContentGenConfig
	GenerateProductContent(ctx context.Context, title, description, platform string) (string, error)
}

func NewContentGeneratorFromEnv() ContentGenerator {
	cfg := LoadContentGenConfig()
	switch cfg.Provider {
	case "stub", "mock":
		return NewStubGenerator(cfg)
	case "openai_compat", "openai-compatible":
		if client := NewOpenAICompatClientFromEnv(); client != nil {
			return client
		}
		return NewStubGenerator(cfg)
	case "nvidia", "":
		if client := NewNVIDIAClientFromEnv(); client != nil {
			return client
		}
		return NewStubGenerator(cfg)
	default:
		return NewStubGenerator(cfg)
	}
}

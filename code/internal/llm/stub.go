package llm

import (
	"context"
	"fmt"
)

type StubGenerator struct {
	cfg ContentGenConfig
}

func NewStubGenerator(cfg ContentGenConfig) *StubGenerator {
	cfg.Enabled = false
	if cfg.Provider == "" {
		cfg.Provider = "stub"
	}
	return &StubGenerator{cfg: cfg}
}

func (s *StubGenerator) ProviderName() string { return "stub" }
func (s *StubGenerator) Enabled() bool        { return false }
func (s *StubGenerator) Config() ContentGenConfig {
	return s.cfg
}
func (s *StubGenerator) GenerateProductContent(ctx context.Context, title, description, platform string) (string, error) {
	_ = ctx
	if description != "" {
		return fmt.Sprintf("基于输入生成内容：%s（%s / %s）", title, description, platform), nil
	}
	return fmt.Sprintf("基于输入生成内容：%s", title), nil
}

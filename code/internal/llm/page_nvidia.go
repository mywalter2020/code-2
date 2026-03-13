package llm

import "context"

type NVIDIAPageGenerator struct {
	client *NVIDIAClient
	cfg    PageGenConfig
}

func NewNVIDIAPageGeneratorFromEnv() *NVIDIAPageGenerator {
	client := NewNVIDIAClientFromEnv()
	if client == nil {
		return nil
	}
	cfg := LoadPageGenConfig()
	client.pageCfg = cfg
	return &NVIDIAPageGenerator{client: client, cfg: cfg}
}

func (g *NVIDIAPageGenerator) ProviderName() string { return "nvidia" }
func (g *NVIDIAPageGenerator) Enabled() bool {
	return g != nil && g.client != nil && g.client.Enabled()
}
func (g *NVIDIAPageGenerator) Config() PageGenConfig { return g.cfg }
func (g *NVIDIAPageGenerator) GeneratePage(ctx context.Context, title, description, platform string) (map[string]any, error) {
	return g.client.GeneratePage(ctx, title, description, platform)
}

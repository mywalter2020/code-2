package types

type AbilityMetadata struct {
	Code        string   `json:"code" yaml:"code"`
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Version     string   `json:"version,omitempty" yaml:"version,omitempty"`
	Enabled     bool     `json:"enabled" yaml:"enabled"`
}

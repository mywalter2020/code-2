package types

type AdapterCredentials struct {
	Platform string            `json:"platform" yaml:"platform"`
	Fields   map[string]string `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type AdapterRequest struct {
	Platform    string         `json:"platform"`
	Action      string         `json:"action"`
	Operator    string         `json:"operator,omitempty"`
	RequestID   string         `json:"request_id,omitempty"`
	ExternalRef string         `json:"external_ref,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

type AdapterResponse struct {
	Platform   string         `json:"platform"`
	Action     string         `json:"action"`
	Status     string         `json:"status"`
	Mode       string         `json:"mode,omitempty"`
	Configured bool           `json:"configured,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

type AdapterHealth struct {
	Platform          string   `json:"platform"`
	Healthy           bool     `json:"healthy"`
	Configured        bool     `json:"configured,omitempty"`
	DryRun            bool     `json:"dry_run,omitempty"`
	BaseURLConfigured bool     `json:"base_url_configured,omitempty"`
	LiveReady         bool     `json:"live_ready,omitempty"`
	Message           string   `json:"message,omitempty"`
	MissingFields     []string `json:"missing_fields,omitempty"`
}

type AdapterDescriptor struct {
	Platform          string   `json:"platform"`
	DisplayName       string   `json:"display_name,omitempty"`
	RequiredFields    []string `json:"required_fields,omitempty"`
	SupportedActions  []string `json:"supported_actions,omitempty"`
	Configured        bool     `json:"configured"`
	DryRun            bool     `json:"dry_run"`
	BaseURLConfigured bool     `json:"base_url_configured,omitempty"`
	LiveReady         bool     `json:"live_ready,omitempty"`
}

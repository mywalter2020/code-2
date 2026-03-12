package types

type AdapterCredentials struct {
	Platform string            `json:"platform" yaml:"platform"`
	Fields   map[string]string `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type AdapterRequest struct {
	Platform string         `json:"platform"`
	Action   string         `json:"action"`
	Operator string         `json:"operator,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
}

type AdapterResponse struct {
	Platform string         `json:"platform"`
	Action   string         `json:"action"`
	Status   string         `json:"status"`
	Data     map[string]any `json:"data,omitempty"`
}

type AdapterHealth struct {
	Platform string `json:"platform"`
	Healthy  bool   `json:"healthy"`
	Message  string `json:"message,omitempty"`
}

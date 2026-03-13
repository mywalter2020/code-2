package types

type TaskFilter struct {
	Status   string `json:"status"`
	Scene    string `json:"scene"`
	Operator string `json:"operator"`
	Platform string `json:"platform"`
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

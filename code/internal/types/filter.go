package types

type TaskFilter struct {
	Status string `json:"status"`
	Scene  string `json:"scene"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

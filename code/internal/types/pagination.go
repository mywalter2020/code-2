package types

type PageResult struct {
	Items  any `json:"items"`
	Count  int `json:"count"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

package meta

type Query struct {
	Type       string   `json:"type"`
	Parameters []string `json:"parameters"`
}

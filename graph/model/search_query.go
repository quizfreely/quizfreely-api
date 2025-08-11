package model

type SearchQuery struct {
	Query   *string `json:"query,omitempty"`
	Subject *string `json:"subject,omitempty"`
}

package model

type Studyset struct {
	ID        *string `json:"id,omitempty"`
	Title     *string `json:"title,omitempty"`
	Private   *bool   `json:"private,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
	UserID      *string   `json:"user_id,omitempty"`
	User      *User   `json:"user,omitempty"`
	Terms     []*Term `json:"terms,omitempty"`
	TermsCount     []*Term `json:"terms_count,omitempty"`
}

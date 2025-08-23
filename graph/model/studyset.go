package model

type Studyset struct {
	ID        *string `json:"id,omitempty"`
	Title     *string `json:"title,omitempty"`
	Private   *bool   `json:"private,omitempty"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
	UserID      *string   `json:"userId,omitempty"`
	User      *User   `json:"user,omitempty"`
	Terms     []*Term `json:"terms,omitempty"`
	TermsCount     []*Term `json:"termsCount,omitempty"`
}

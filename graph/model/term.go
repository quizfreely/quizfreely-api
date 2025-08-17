package model

type Term struct {
	ID        *string       `json:"id,omitempty"`
	Term      *string       `json:"term,omitempty"`
	Def       *string       `json:"def,omitempty"`
	StudysetID *string
	SortOrder *int32        `json:"sort_order,omitempty"`
	Progress  *TermProgress `json:"progress,omitempty"`
	CreatedAt *string       `json:"created_at,omitempty"`
	UpdatedAt *string       `json:"updated_at,omitempty"`
}

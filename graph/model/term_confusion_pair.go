package model

type TermConfusionPair struct {
	ID             *string     `json:"id,omitempty"`
	TermID           *string       `json:"term_id,omitempty"`
	ConfusedTermID   *string       `json:"confused_term_id,omitempty"`
	ConfusedTerm   *Term       `json:"confused_term,omitempty"`
	AnsweredWith   *AnswerWith `json:"answered_with,omitempty"`
	ConfusedCount  *int32      `json:"confused_count,omitempty"`
	LastConfusedAt *string     `json:"last_confused_at,omitempty"`
}

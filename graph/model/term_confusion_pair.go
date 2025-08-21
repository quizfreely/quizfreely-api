package model

type TermConfusionPair struct {
	ID             *string     `json:"id,omitempty"`
	Term           *Term       `json:"term,omitempty"`
	ConfusedTerm   *Term       `json:"confused_term,omitempty"`
	AnsweredWith   *AnswerWith `json:"answered_with,omitempty"`
	ConfusedCount  *int32      `json:"confused_count,omitempty"`
	LastConfusedAt *string     `json:"last_confused_at,omitempty"`
}

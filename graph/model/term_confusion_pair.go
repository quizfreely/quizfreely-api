package model

type TermConfusionPair struct {
	ID             *string     `json:"id,omitempty"`
	TermID           *string       `json:"termId,omitempty"`
	Term *Term `json:"term,omitempty"`
	ConfusedTermID   *string       `json:"confusedTermId,omitempty"`
	ConfusedTerm   *Term       `json:"confusedTerm,omitempty"`
	AnsweredWith   *AnswerWith `json:"answeredWith,omitempty"`
	ConfusedCount  *int32      `json:"confusedCount,omitempty"`
	LastConfusedAt *string     `json:"lastConfusedAt,omitempty"`
}

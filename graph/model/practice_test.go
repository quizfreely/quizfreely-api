package model

type PracticeTest struct {
	ID               *string     `json:"id,omitempty"`
	StudysetID       *string     `json:"studyset_id,omitempty"`
	Timestamp        *string     `json:"timestamp,omitempty"`
	QuestionsCorrect *int32      `json:"questions_correct,omitempty"`
	QuestionsTotal   *int32      `json:"questions_total,omitempty"`
	Questions        []*Question `json:"questions,omitempty"`
}

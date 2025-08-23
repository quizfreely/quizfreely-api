package model

type PracticeTest struct {
	ID               *string     `json:"id,omitempty"`
	StudysetID       *string     `json:"studysetId,omitempty"`
	Timestamp        *string     `json:"timestamp,omitempty"`
	QuestionsCorrect *int32      `json:"questionsCorrect,omitempty"`
	QuestionsTotal   *int32      `json:"questionsTotal,omitempty"`
	Questions        []*Question `json:"questions,omitempty"`
}

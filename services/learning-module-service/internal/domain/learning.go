package domain

type LessonType string

const (
	LessonTypeMultipleChoice LessonType = "multiple_choice"
	LessonTypeCodeChallenge  LessonType = "code_challenge"
	LessonTypeTheory         LessonType = "theory"
	LessonTypeDocument       LessonType = "document"
)

type AnswerOption struct {
	ID        string
	Text      string
	IsCorrect bool
}

type Lesson struct {
	ID             string
	ChapterID      string
	Order          int
	Title          string
	Description    string
	Type           LessonType
	Options        []AnswerOption
	StarterCode    string
	SolutionCode   string
	TestCode       string
	ExpectedOutput string
	Language       string
	DocumentURL    string
}

type Chapter struct {
	ID       string
	CourseID string
	Order    int
	Title    string
	Lessons  []Lesson
}

type Course struct {
	ID            string
	Title         string
	Description   string
	Language      string
	Category      string
	Difficulty    string
	Chapters      []Chapter
	ChaptersCount int
	TasksCount    int
	LessonIDs     []string
}

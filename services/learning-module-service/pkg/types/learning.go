package types

// CourseListItem is returned when listing all courses.
type CourseListItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Language      string   `json:"language,omitempty"`
	Category      string   `json:"category,omitempty"`
	Difficulty    string   `json:"difficulty,omitempty"`
	ChaptersCount int      `json:"chapters_count"`
	TasksCount    int      `json:"tasks_count"`
	LessonIDs     []string `json:"lesson_ids"`
}

// CourseDetail includes chapters and lesson summaries.
type CourseDetail struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Language    string           `json:"language,omitempty"`
	Category    string           `json:"category,omitempty"`
	Difficulty  string           `json:"difficulty,omitempty"`
	Chapters    []ChapterSummary `json:"chapters"`
}

type ChapterSummary struct {
	ID      string          `json:"id"`
	Order   int             `json:"order"`
	Title   string          `json:"title"`
	Lessons []LessonSummary `json:"lessons"`
}

type LessonSummary struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// LessonDetail is the full lesson content sent to the learner.
// Correct answers are never included here.
type LessonDetail struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Options     []AnswerOption `json:"options,omitempty"`
	StarterCode string         `json:"starter_code,omitempty"`
	Language    string         `json:"language,omitempty"`
	DocumentURL string         `json:"document_url,omitempty"`
}

// AnswerOption is a multiple-choice option without the correct-answer flag.
type AnswerOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// SubmitAnswerRequest is the request body for submitting a lesson answer.
type SubmitAnswerRequest struct {
	SelectedOptionID string `json:"selected_option_id,omitempty"` // multiple_choice
	Code             string `json:"code,omitempty"`               // code_challenge
}

// SubmitAnswerResponse is the result of a submission.
type SubmitAnswerResponse struct {
	Correct       bool   `json:"correct"`
	Feedback      string `json:"feedback"`
	ConsoleOutput string `json:"console_output,omitempty"`
}

// RunCodeRequest is the request body for running code without submitting.
type RunCodeRequest struct {
	Code string `json:"code"`
}

// RunCodeResponse is the result of a code run.
type RunCodeResponse struct {
	ConsoleOutput string `json:"console_output"`
	ExitCode      int    `json:"exit_code"`
}

// CreateCourseRequest is the request body for creating a course.
type CreateCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language,omitempty"`
	Category    string `json:"category,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
}

// CreateChapterRequest is the request body for creating a chapter.
type CreateChapterRequest struct {
	Order int    `json:"order"`
	Title string `json:"title"`
}

// CreateLessonOption is an option when creating a multiple-choice lesson.
type CreateLessonOption struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

// CreateLessonRequest is the request body for creating a lesson.
type CreateLessonRequest struct {
	Order          int                  `json:"order"`
	Title          string               `json:"title"`
	Description    string               `json:"description"`
	Type           string               `json:"type"` // "multiple_choice" or "code_challenge"
	Language       string               `json:"language,omitempty"`
	StarterCode    string               `json:"starter_code,omitempty"`
	SolutionCode   string               `json:"solution_code,omitempty"`
	TestCode       string               `json:"test_code,omitempty"`
	ExpectedOutput string               `json:"expected_output,omitempty"`
	Options        []CreateLessonOption `json:"options,omitempty"`
	DocumentURL    string               `json:"document_url,omitempty"`
}

// RequestHelpResponse is the result of asking for help.
type RequestHelpResponse struct {
	Hint      string `json:"hint"`
	HelpCount int    `json:"help_count"`
}

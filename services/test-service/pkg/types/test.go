package types

type TestModel struct {
	ID             string  `json:"id"`
	CourseID       string  `json:"course_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	IsVisible      bool    `json:"is_visible"`
	Status         string  `json:"status"`
	OpenDate       *string `json:"open_date,omitempty"` // RFC3339
	Language       string  `json:"language"`
	StarterCode    string  `json:"starter_code"`
	SolutionCode   string  `json:"solution_code"`
	TestCode       string  `json:"test_code"`
	ExpectedOutput string  `json:"expected_output"`
	GroupID        *string `json:"group_id,omitempty"`
}

type CreateTestRequest struct {
	CourseID       string  `json:"course_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	IsVisible      bool    `json:"is_visible"`
	Status         string  `json:"status"`
	OpenDate       *string `json:"open_date,omitempty"`
	Language       string  `json:"language"`
	StarterCode    string  `json:"starter_code"`
	SolutionCode   string  `json:"solution_code"`
	TestCode       string  `json:"test_code"`
	ExpectedOutput string  `json:"expected_output"`
	GroupID        *string `json:"group_id,omitempty"`
}

type UpdateTestRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	IsVisible      bool    `json:"is_visible"`
	Status         string  `json:"status"`
	OpenDate       *string `json:"open_date,omitempty"`
	Language       string  `json:"language"`
	StarterCode    string  `json:"starter_code"`
	SolutionCode   string  `json:"solution_code"`
	TestCode       string  `json:"test_code"`
	ExpectedOutput string  `json:"expected_output"`
	GroupID        *string `json:"group_id,omitempty"`
}

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SubmitTestRequest struct {
	Code  string `json:"code"`
	Stdin string `json:"stdin"`
}

type SubmitTestResponse struct {
	Correct  bool   `json:"correct"`
	Feedback string `json:"feedback"`
	Output   string `json:"output"`
}

type RunTestResponse struct {
	ConsoleOutput string `json:"console_output"`
	ExitCode      int    `json:"exit_code"`
}

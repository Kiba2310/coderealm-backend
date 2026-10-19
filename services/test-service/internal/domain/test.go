package domain

type Test struct {
	ID             string
	CourseID       string
	Title          string
	Description    string
	IsVisible      bool
	Status         string
	OpenDate       *string // RFC3339 formatted date string or nil
	Language       string
	StarterCode    string
	SolutionCode   string
	TestCode       string
	ExpectedOutput string
	GroupID        *string // nil = global test, set = group-specific test
}

type TestRepository interface {
	GetTests() ([]*Test, error)
	GetTestByID(id string) (*Test, error)
	CreateTest(t *Test) error
	UpdateTest(t *Test) error
	DeleteTest(id string) error
	GetTestsByGroupID(groupID string) ([]*Test, error)
}

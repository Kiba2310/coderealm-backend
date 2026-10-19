package service

import (
	"fmt"
	"strings"

	"BootDevClone/services/test-service/internal/domain"
	"BootDevClone/services/test-service/internal/infrastructure/piston"
	"BootDevClone/services/test-service/pkg/types"

	"github.com/google/uuid"
)

type TestService struct {
	repo         domain.TestRepository
	pistonClient *piston.Client
}

func NewTestService(repo domain.TestRepository, pistonClient *piston.Client) *TestService {
	return &TestService{
		repo:         repo,
		pistonClient: pistonClient,
	}
}

func domainToModel(t *domain.Test) types.TestModel {
	return types.TestModel{
		ID:             t.ID,
		CourseID:       t.CourseID,
		Title:          t.Title,
		Description:    t.Description,
		IsVisible:      t.IsVisible,
		Status:         t.Status,
		OpenDate:       t.OpenDate,
		Language:       t.Language,
		StarterCode:    t.StarterCode,
		SolutionCode:   t.SolutionCode,
		TestCode:       t.TestCode,
		ExpectedOutput: t.ExpectedOutput,
		GroupID:        t.GroupID,
	}
}

func (s *TestService) GetTestByID(id string) (*types.TestModel, error) {
	t, err := s.repo.GetTestByID(id)
	if err != nil {
		return nil, err
	}
	model := domainToModel(t)
	return &model, nil
}

func (s *TestService) SubmitTest(id string, req types.SubmitTestRequest) (*types.SubmitTestResponse, error) {
	test, err := s.repo.GetTestByID(id)
	if err != nil {
		return nil, err
	}

	resp, err := s.pistonClient.ExecuteCode(test.Language, req.Code, test.TestCode, req.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	output := resp.Run.Stdout + resp.Run.Stderr
	correct := true
	feedback := "Correct! Well done."

	if test.TestCode != "" && resp.Run.Code != 0 {
		correct = false
		feedback = "Tests failed. Check the console output."
	}

	if test.ExpectedOutput != "" {
		if strings.TrimSpace(resp.Run.Stdout) != strings.TrimSpace(test.ExpectedOutput) {
			correct = false
			feedback = "Output did not match the expected output."
		}
	}

	if test.TestCode == "" && test.ExpectedOutput == "" && resp.Run.Code != 0 {
		correct = false
		feedback = "Execution failed. Check the console output."
	}

	return &types.SubmitTestResponse{
		Correct:  correct,
		Feedback: feedback,
		Output:   output,
	}, nil
}

func (s *TestService) RunTest(id string, req types.SubmitTestRequest) (*types.RunTestResponse, error) {
	test, err := s.repo.GetTestByID(id)
	if err != nil {
		return nil, err
	}

	// For simple run, do NOT pass test.TestCode, only run the user's code
	resp, err := s.pistonClient.ExecuteCode(test.Language, req.Code, "", req.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	return &types.RunTestResponse{
		ConsoleOutput: resp.Run.Stdout + resp.Run.Stderr,
		ExitCode:      resp.Run.Code,
	}, nil
}

func (s *TestService) GetTests() ([]types.TestModel, error) {
	tests, err := s.repo.GetTests()
	if err != nil {
		return nil, err
	}
	res := make([]types.TestModel, 0, len(tests))
	for _, t := range tests {
		res = append(res, domainToModel(t))
	}
	return res, nil
}

func (s *TestService) GetTestsByGroupID(groupID string) ([]types.TestModel, error) {
	tests, err := s.repo.GetTestsByGroupID(groupID)
	if err != nil {
		return nil, err
	}
	res := make([]types.TestModel, 0, len(tests))
	for _, t := range tests {
		res = append(res, domainToModel(t))
	}
	return res, nil
}

func (s *TestService) CreateTest(req types.CreateTestRequest) (*types.TestModel, error) {
	id := uuid.New().String()
	t := &domain.Test{
		ID:             id,
		CourseID:       req.CourseID,
		Title:          req.Title,
		Description:    req.Description,
		IsVisible:      req.IsVisible,
		Status:         req.Status,
		OpenDate:       req.OpenDate,
		Language:       req.Language,
		StarterCode:    req.StarterCode,
		SolutionCode:   req.SolutionCode,
		TestCode:       req.TestCode,
		ExpectedOutput: req.ExpectedOutput,
		GroupID:        req.GroupID,
	}
	if err := s.repo.CreateTest(t); err != nil {
		return nil, err
	}
	model := domainToModel(t)
	return &model, nil
}

func (s *TestService) UpdateTest(id string, req types.UpdateTestRequest) error {
	t := &domain.Test{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		IsVisible:      req.IsVisible,
		Status:         req.Status,
		OpenDate:       req.OpenDate,
		Language:       req.Language,
		StarterCode:    req.StarterCode,
		SolutionCode:   req.SolutionCode,
		TestCode:       req.TestCode,
		ExpectedOutput: req.ExpectedOutput,
		GroupID:        req.GroupID,
	}
	return s.repo.UpdateTest(t)
}

func (s *TestService) DeleteTest(id string) error {
	return s.repo.DeleteTest(id)
}

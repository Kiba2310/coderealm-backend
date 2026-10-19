package service

import (
	"BootDevClone/services/learning-module-service/internal/domain"
	"BootDevClone/services/learning-module-service/internal/infrastructure/piston"
	"BootDevClone/services/learning-module-service/pkg/types"
	"BootDevClone/shared/cache"
	"BootDevClone/shared/rabbitmq"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LearningRepository interface {
	ListCourses(lang string) ([]*domain.Course, error)
	GetCourse(id string, lang string) (*domain.Course, error)
	GetLesson(courseID, chapterID, lessonID string, lang string) (*domain.Lesson, error)
	CreateCourse(course *domain.Course) error
	CreateChapter(chapter *domain.Chapter) error
	CreateLesson(lesson *domain.Lesson) error
	UpdateCourse(course *domain.Course) error
	DeleteCourse(id string) error
	UpdateChapter(chapter *domain.Chapter) error
	DeleteChapter(courseID, chapterID string) error
	UpdateLesson(lesson *domain.Lesson) error
	DeleteLesson(courseID, chapterID, lessonID string) error

	IncrementLessonHelp(userID, lessonID string) (int, error)
	GetLessonHelpCount(userID, lessonID string) (int, error)
}

type LearningService struct {
	repo            LearningRepository
	pistonClient    *piston.Client
	rmqClient       *rabbitmq.Client
	cacheClient     *cache.RedisClient
	cachedLanguages []string
	langMutex       sync.RWMutex
}

func NewLearningService(repo LearningRepository, pistonClient *piston.Client, rmqClient *rabbitmq.Client, cacheClient *cache.RedisClient) *LearningService {
	return &LearningService{
		repo:         repo,
		pistonClient: pistonClient,
		rmqClient:    rmqClient,
		cacheClient:  cacheClient,
	}
}

func (s *LearningService) GetLanguages() ([]string, error) {
	s.langMutex.RLock()
	if len(s.cachedLanguages) > 0 {
		langs := s.cachedLanguages
		s.langMutex.RUnlock()
		return langs, nil
	}
	s.langMutex.RUnlock()

	packages, err := s.pistonClient.GetPackages()
	if err != nil {
		// If piston is completely down, return some sensible defaults rather than failing the whole page
		fmt.Printf("Warning: failed to get packages from Piston API: %v\n", err)
		return []string{"python", "javascript", "go", "java", "c", "cpp"}, nil
	}

	// Use a map to ensure uniqueness
	langMap := make(map[string]bool)
	for _, pkg := range packages {
		langMap[pkg.Language] = true
	}

	var languages []string
	for lang := range langMap {
		languages = append(languages, lang)
	}

	s.langMutex.Lock()
	s.cachedLanguages = languages
	s.langMutex.Unlock()

	return languages, nil
}

func (s *LearningService) ListCourses(lang string) ([]types.CourseListItem, error) {
	cacheKey := fmt.Sprintf("courses:list:%s", lang)
	var cachedResult []types.CourseListItem

	if s.cacheClient != nil {
		if err := s.cacheClient.Get(context.Background(), cacheKey, &cachedResult); err == nil {
			return cachedResult, nil
		}
	}

	courses, err := s.repo.ListCourses(lang)
	if err != nil {
		return nil, err
	}
	result := make([]types.CourseListItem, 0, len(courses))
	for _, c := range courses {
		result = append(result, types.CourseListItem{
			ID:            c.ID,
			Title:         c.Title,
			Description:   c.Description,
			Language:      c.Language,
			Category:      c.Category,
			Difficulty:    c.Difficulty,
			ChaptersCount: c.ChaptersCount,
			TasksCount:    c.TasksCount,
			LessonIDs:     c.LessonIDs,
		})
	}

	if s.cacheClient != nil {
		s.cacheClient.Set(context.Background(), cacheKey, result, 1*time.Hour)
	}

	return result, nil
}

func (s *LearningService) GetCourseDetail(courseID string, lang string) (*types.CourseDetail, error) {
	cacheKey := fmt.Sprintf("course:%s:%s", courseID, lang)
	var cachedDetail types.CourseDetail

	if s.cacheClient != nil {
		if err := s.cacheClient.Get(context.Background(), cacheKey, &cachedDetail); err == nil {
			return &cachedDetail, nil
		}
	}

	course, err := s.repo.GetCourse(courseID, lang)
	if err != nil {
		return nil, err
	}
	chapters := make([]types.ChapterSummary, 0, len(course.Chapters))
	for _, ch := range course.Chapters {
		lessons := make([]types.LessonSummary, 0, len(ch.Lessons))
		for _, l := range ch.Lessons {
			lessons = append(lessons, types.LessonSummary{
				ID:    l.ID,
				Order: l.Order,
				Title: l.Title,
				Type:  string(l.Type),
			})
		}
		chapters = append(chapters, types.ChapterSummary{
			ID:      ch.ID,
			Order:   ch.Order,
			Title:   ch.Title,
			Lessons: lessons,
		})
	}
	detail := &types.CourseDetail{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Language:    course.Language,
		Category:    course.Category,
		Difficulty:  course.Difficulty,
		Chapters:    chapters,
	}

	if s.cacheClient != nil {
		s.cacheClient.Set(context.Background(), cacheKey, detail, 1*time.Hour)
	}

	return detail, nil
}

func (s *LearningService) GetLesson(courseID, chapterID, lessonID string, lang string) (*types.LessonDetail, error) {
	cacheKey := fmt.Sprintf("course:%s:lesson:%s:%s", courseID, lessonID, lang)
	var cachedDetail types.LessonDetail

	if s.cacheClient != nil {
		if err := s.cacheClient.Get(context.Background(), cacheKey, &cachedDetail); err == nil {
			return &cachedDetail, nil
		}
	}

	lesson, err := s.repo.GetLesson(courseID, chapterID, lessonID, lang)
	if err != nil {
		return nil, err
	}
	detail := &types.LessonDetail{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Description: lesson.Description,
		Type:        string(lesson.Type),
		Language:    lesson.Language,
		StarterCode: lesson.StarterCode,
		DocumentURL: lesson.DocumentURL,
	}
	if lesson.Type == domain.LessonTypeMultipleChoice {
		opts := make([]types.AnswerOption, 0, len(lesson.Options))
		for _, o := range lesson.Options {
			opts = append(opts, types.AnswerOption{
				ID:   o.ID,
				Text: o.Text,
			})
		}
		detail.Options = opts
	}

	if s.cacheClient != nil {
		s.cacheClient.Set(context.Background(), cacheKey, detail, 1*time.Hour)
	}

	return detail, nil
}

func (s *LearningService) RunCode(courseID, chapterID, lessonID string, req types.RunCodeRequest, lang string) (*types.RunCodeResponse, error) {
	lesson, err := s.repo.GetLesson(courseID, chapterID, lessonID, lang)
	if err != nil {
		return nil, err
	}
	if lesson.Type != domain.LessonTypeCodeChallenge {
		return nil, errors.New("not a code challenge")
	}

	resp, err := s.pistonClient.ExecuteCode(lesson.Language, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	return &types.RunCodeResponse{
		ConsoleOutput: resp.Run.Stdout + resp.Run.Stderr,
		ExitCode:      resp.Run.Code,
	}, nil
}

type LessonCompletedEvent struct {
	UserID     string `json:"user_id"`
	LessonID   string `json:"lesson_id"`
	LessonType string `json:"lesson_type"`
	HelpCount  int    `json:"help_count"`
}

func (s *LearningService) SubmitAnswer(courseID, chapterID, lessonID, userID string, req types.SubmitAnswerRequest, lang string) (*types.SubmitAnswerResponse, error) {
	lesson, err := s.repo.GetLesson(courseID, chapterID, lessonID, lang)
	if err != nil {
		return nil, err
	}

	helpCount, _ := s.repo.GetLessonHelpCount(userID, lessonID)

	publishCompletion := func() {
		if s.rmqClient != nil && userID != "" {
			event := LessonCompletedEvent{
				UserID:     userID,
				LessonID:   lessonID,
				LessonType: string(lesson.Type),
				HelpCount:  helpCount,
			}
			err := s.rmqClient.Publish("lesson.completed", event)
			if err != nil {
				fmt.Printf("Failed to publish lesson.completed event: %v\n", err)
			}
		}
	}

	switch lesson.Type {
	case domain.LessonTypeTheory, domain.LessonTypeDocument:
		publishCompletion()
		msg := "Lesson completed!"
		if lang == "hu" {
			msg = "Lecke teljesítve!"
		}
		return &types.SubmitAnswerResponse{Correct: true, Feedback: msg}, nil
	case domain.LessonTypeMultipleChoice:
		for _, opt := range lesson.Options {
			if opt.ID == req.SelectedOptionID {
				if opt.IsCorrect {
					publishCompletion()
					msg := "Correct! Well done."
					if lang == "hu" {
						msg = "Helyes! Szép munka."
					}
					return &types.SubmitAnswerResponse{Correct: true, Feedback: msg}, nil
				}
				msg := "Not quite. Try again!"
				if lang == "hu" {
					msg = "Nem egészen. Próbáld újra!"
				}
				return &types.SubmitAnswerResponse{Correct: false, Feedback: msg}, nil
			}
		}
		if lang == "hu" {
			return nil, errors.New("érvénytelen válasz opció")
		}
		return nil, errors.New("invalid option id")
	case domain.LessonTypeCodeChallenge:
		codeToRun := req.Code
		if lesson.TestCode != "" {
			codeToRun = req.Code + "\n" + lesson.TestCode
		}

		resp, err := s.pistonClient.ExecuteCode(lesson.Language, codeToRun)
		if err != nil {
			return nil, fmt.Errorf("failed to execute code: %w", err)
		}

		output := resp.Run.Stdout + resp.Run.Stderr
		correct := true
		feedback := "Correct! Well done."
		if lang == "hu" {
			feedback = "Helyes! Szép munka."
		}

		// 1. If TestCode is present, check exit code
		if lesson.TestCode != "" && resp.Run.Code != 0 {
			correct = false
			feedback = "Tests failed. Check the console output."
			if lang == "hu" {
				feedback = "A tesztek elbuktak. Ellenőrizd a konzol kimenetét."
			}
		}

		// 2. If ExpectedOutput is present, check string match
		if lesson.ExpectedOutput != "" {
			// We trim space to be forgiving with trailing newlines
			if strings.TrimSpace(resp.Run.Stdout) != strings.TrimSpace(lesson.ExpectedOutput) {
				correct = false
				feedback = "Output did not match the expected output."
				if lang == "hu" {
					feedback = "A kimenet nem egyezik az elvárt kimenettel."
				}
			}
		}

		// If both are missing, it just returns true if it ran (fallback)
		if lesson.TestCode == "" && lesson.ExpectedOutput == "" && resp.Run.Code != 0 {
			correct = false
			feedback = "Execution failed. Check the console output."
			if lang == "hu" {
				feedback = "Futtatás sikertelen. Ellenőrizd a konzol kimenetét."
			}
		}

		if correct {
			publishCompletion()
		}

		return &types.SubmitAnswerResponse{
			Correct:       correct,
			Feedback:      feedback,
			ConsoleOutput: output,
		}, nil
	}
	if lang == "hu" {
		return nil, errors.New("nem támogatott lecke típus")
	}
	return nil, errors.New("unsupported lesson type")
}

func (s *LearningService) RequestHelp(courseID, chapterID, lessonID, userID, lang string) (*types.RequestHelpResponse, error) {
	count, err := s.repo.GetLessonHelpCount(userID, lessonID)
	if err != nil {
		return nil, err
	}
	if count >= 3 {
		if lang == "hu" {
			return nil, errors.New("elérted a maximális segítségkérési lehetőséget ehhez a leckéhez")
		}
		return nil, errors.New("maximum help limit reached for this lesson")
	}

	lesson, err := s.repo.GetLesson(courseID, chapterID, lessonID, lang)
	if err != nil {
		return nil, err
	}

	if lesson.Type == domain.LessonTypeTheory || lesson.Type == domain.LessonTypeDocument {
		if lang == "hu" {
			return nil, errors.New("elméleti vagy dokumentum leckékhez nem kérhető segítség")
		}
		return nil, errors.New("theory/document lessons do not support help")
	}

	newCount, err := s.repo.IncrementLessonHelp(userID, lessonID)
	if err != nil {
		return nil, err
	}

	var hint string
	switch lesson.Type {
	case domain.LessonTypeMultipleChoice:
		var incorrectOptions []domain.AnswerOption
		for _, opt := range lesson.Options {
			if !opt.IsCorrect {
				incorrectOptions = append(incorrectOptions, opt)
			}
		}
		if len(incorrectOptions) > 0 {
			idx := (newCount - 1) % len(incorrectOptions)
			hint = "Egy kis segítség: A helyes válasz BIZTOSAN NEM a(z) '" + incorrectOptions[idx].Text + "'."
		} else {
			hint = "Sajnos ehhez a kérdéshez nincs több elérhető segítség."
		}
	case domain.LessonTypeCodeChallenge:
		if newCount == 1 {
			hint = "Tipp: Olvasd át alaposan a feladat leírását! Ellenőrizd a változók neveit és a kért kimenetet."
		} else {
			if lesson.SolutionCode == "" {
				hint = "Sajnos a tanár nem adott meg hivatalos megoldást ehhez a feladathoz."
			} else {
				percentage := 0.30
				if newCount >= 3 {
					percentage = 0.75
				}

				runes := []rune(lesson.SolutionCode)
				revealCount := int(float64(len(runes)) * percentage)
				if revealCount == 0 && len(runes) > 0 {
					revealCount = 1
				}

				revealed := string(runes[:revealCount])
				hint = fmt.Sprintf("Itt egy részlet a megoldásból:\n\n```%s\n%s...\n```", lesson.Language, revealed)
			}
		}
	}

	return &types.RequestHelpResponse{
		Hint:      hint,
		HelpCount: newCount,
	}, nil
}

func (s *LearningService) invalidateCourseCache() {
	if s.cacheClient != nil {
		s.cacheClient.DeletePattern(context.Background(), "course*")
	}
}

func (s *LearningService) CreateCourse(req types.CreateCourseRequest) (*types.CourseListItem, error) {
	id := uuid.New().String()
	course := &domain.Course{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Language:    req.Language,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
	}

	if course.Language != "" {
		// Asynchronously ensure runtime is installed so it doesn't block if it's very slow,
		// or we can do it synchronously. The user said it might take a few seconds so let's do it synchronously
		// to guarantee it's ready, and the frontend will show a spinner.
		if err := s.pistonClient.EnsureRuntime(course.Language); err != nil {
			fmt.Printf("Warning: EnsureRuntime failed for %s: %v\n", course.Language, err)
			// We continue anyway, so the course is created even if piston is down temporarily
		}
	}

	if err := s.repo.CreateCourse(course); err != nil {
		return nil, err
	}
	s.invalidateCourseCache()
	return &types.CourseListItem{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
	}, nil
}

func (s *LearningService) CreateChapter(courseID string, req types.CreateChapterRequest) (*types.ChapterSummary, error) {
	id := uuid.New().String()
	chapter := &domain.Chapter{
		ID:       id,
		CourseID: courseID,
		Order:    req.Order,
		Title:    req.Title,
	}
	if err := s.repo.CreateChapter(chapter); err != nil {
		return nil, err
	}
	s.invalidateCourseCache()
	return &types.ChapterSummary{
		ID:      id,
		Order:   req.Order,
		Title:   req.Title,
		Lessons: []types.LessonSummary{},
	}, nil
}

func (s *LearningService) CreateLesson(courseID, chapterID string, req types.CreateLessonRequest) (*types.LessonDetail, error) {
	id := uuid.New().String()
	lesson := &domain.Lesson{
		ID:           id,
		ChapterID:    chapterID,
		Order:        req.Order,
		Title:        req.Title,
		Description:  req.Description,
		Type:         domain.LessonType(req.Type),
		Language:     req.Language,
		StarterCode:  req.StarterCode,
		SolutionCode: req.SolutionCode,
		TestCode:     req.TestCode,
		DocumentURL:  req.DocumentURL,
	}

	if lesson.Type == domain.LessonTypeMultipleChoice {
		for _, o := range req.Options {
			lesson.Options = append(lesson.Options, domain.AnswerOption{
				ID:        uuid.New().String(),
				Text:      o.Text,
				IsCorrect: o.IsCorrect,
			})
		}
	}

	if err := s.repo.CreateLesson(lesson); err != nil {
		return nil, err
	}
	s.invalidateCourseCache()

	res := &types.LessonDetail{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Language:    req.Language,
		StarterCode: req.StarterCode,
		DocumentURL: req.DocumentURL,
	}

	if lesson.Type == domain.LessonTypeMultipleChoice {
		opts := make([]types.AnswerOption, 0, len(lesson.Options))
		for _, o := range lesson.Options {
			opts = append(opts, types.AnswerOption{
				ID:   o.ID,
				Text: o.Text,
			})
		}
		res.Options = opts
	}

	return res, nil
}

// ==================== Update/Delete Methods ====================

func (s *LearningService) UpdateCourse(courseID string, req types.CreateCourseRequest) error {
	course := &domain.Course{
		ID:          courseID,
		Title:       req.Title,
		Description: req.Description,
		Language:    req.Language,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
	}
	err := s.repo.UpdateCourse(course)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

func (s *LearningService) DeleteCourse(courseID string) error {
	err := s.repo.DeleteCourse(courseID)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

func (s *LearningService) UpdateChapter(courseID, chapterID string, req types.CreateChapterRequest) error {
	chapter := &domain.Chapter{
		ID:       chapterID,
		CourseID: courseID,
		Order:    req.Order,
		Title:    req.Title,
	}
	err := s.repo.UpdateChapter(chapter)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

func (s *LearningService) DeleteChapter(courseID, chapterID string) error {
	err := s.repo.DeleteChapter(courseID, chapterID)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

func (s *LearningService) UpdateLesson(courseID, chapterID, lessonID string, req types.CreateLessonRequest) error {
	lesson := &domain.Lesson{
		ID:             lessonID,
		ChapterID:      chapterID,
		Order:          req.Order,
		Title:          req.Title,
		Description:    req.Description,
		Type:           domain.LessonType(req.Type),
		Language:       req.Language,
		StarterCode:    req.StarterCode,
		SolutionCode:   req.SolutionCode,
		TestCode:       req.TestCode,
		ExpectedOutput: req.ExpectedOutput,
		DocumentURL:    req.DocumentURL,
	}
	err := s.repo.UpdateLesson(lesson)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

func (s *LearningService) DeleteLesson(courseID, chapterID, lessonID string) error {
	err := s.repo.DeleteLesson(courseID, chapterID, lessonID)
	if err == nil {
		s.invalidateCourseCache()
	}
	return err
}

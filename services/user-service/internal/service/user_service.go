package service

import (
	"errors"
	"strings"

	"BootDevClone/services/user-service/internal/domain"
	"BootDevClone/services/user-service/internal/infrastructure/piston"
	"BootDevClone/services/user-service/internal/infrastructure/repository"
	"BootDevClone/shared/cache"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	repo         *repository.UserRepository
	cacheClient  *cache.RedisClient
	pistonClient *piston.Client
}

func NewUserService(repo *repository.UserRepository, _ string, cacheClient *cache.RedisClient) *UserService {
	// _ represents the legacy jwtSecret param which is no longer needed
	return &UserService{
		repo:         repo,
		cacheClient:  cacheClient,
		pistonClient: piston.NewClient(),
	}
}

func (s *UserService) SyncUser(keycloakID, email, rolesStr string) (*domain.User, error) {
	// Find user by Keycloak ID
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		// User does not exist, create them
		parts := strings.Split(email, "@")
		nn := "Diák"
		if len(parts) > 0 {
			nn = parts[0]
		}

		user = &domain.User{
			ID:         uuid.New(),
			KeycloakID: keycloakID,
			Email:      email,
			Nickname:   &nn,
		}

		if err := s.repo.CreateUser(user); err != nil {
			return nil, err
		}
	} else if user.Email != email {
		// Just in case email changes in Keycloak, we could update it, but skip for now
	}

	// Roles management
	roles := strings.Split(rolesStr, ",")
	isAdmin := false
	for _, r := range roles {
		if strings.EqualFold(strings.TrimSpace(r), "admin") {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		_ = s.repo.AssignRole(user.ID, "admin")
	} else {
		existingRoles, _ := s.repo.GetUserRoles(user.ID)
		if len(existingRoles) == 0 {
			_ = s.repo.AssignRole(user.ID, "student")
		}
	}

	s.repo.TouchUserActivity(user.ID)

	return s.repo.GetUserByID(user.ID)
}

func (s *UserService) GetUserByID(id string) (*domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(uid)
}

func (s *UserService) ConsumeHelperItem(userID uuid.UUID, lessonID string) (bool, error) {
	return s.repo.ConsumeHelperItem(userID, lessonID)
}

// CompleteLesson records a lesson completion and awards points based on lesson type.
// Returns (pointsAwarded, totalPoints, helperUsed, error).
func (s *UserService) CompleteLesson(keycloakID, lessonID, lessonType string, helpCount int) (int, int, bool, error) {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return 0, 0, false, err
	}

	pointsToAward := 10 // default
	switch lessonType {
	case "theory":
		pointsToAward = 1
	case "multiple_choice":
		pointsToAward = 5
	case "code_challenge":
		pointsToAward = 10
	}

	helperUsed := false
	if helpCount > 0 {
		usedHelper, _ := s.repo.CheckIfHelperUsedOnLesson(user.ID, lessonID)
		if usedHelper {
			helpCount--
			helperUsed = true
		}
	}

	if helpCount > 0 {
		pointsToAward = int(float64(pointsToAward) * float64(100-(helpCount*20)) / 100.0)
		if pointsToAward < 0 {
			pointsToAward = 0
		}
	}

	hasAmulet, _ := s.repo.HasActiveXPAmulet(user.ID)
	if hasAmulet {
		pointsToAward = int(float64(pointsToAward) * 1.1)
	}

	awarded, totalPoints, err := s.repo.CompleteLessonAndAwardPoints(user.ID, lessonID, lessonType, pointsToAward)
	if err != nil {
		return 0, 0, false, err
	}

	if awarded {
		// Synchronously check and award achievements if conditions are met
		totalCount, _ := s.repo.GetUserTotalCompletedLessonsCount(user.ID)
		if totalCount == 1 {
			_ = s.repo.AddAchievement(user.ID, "first_lesson")
		}

		if lessonType == "code_challenge" {
			codeCount, _ := s.repo.GetUserCodeChallengeCount(user.ID)
			if codeCount == 10 {
				_ = s.repo.AddAchievement(user.ID, "code_warrior")
			}
		}

		// Hardcoded fallback for "master" achievement if total lessons reaches e.g. 20
		// In a real system, we'd query learning-module-service for total available lessons.
		if totalCount >= 20 {
			_ = s.repo.AddAchievement(user.ID, "master")
		}
	}

	if awarded {
		return pointsToAward, totalPoints, helperUsed, nil
	}
	return 0, totalPoints, false, nil
}

// ResetProgress deletes all completed lessons and points for a user.
func (s *UserService) ResetProgress(keycloakID string) error {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return err
	}
	return s.repo.ResetProgress(user.ID)
}

// UpdateProfile updates the editable profile fields of a user.
func (s *UserService) UpdateProfile(keycloakID string, nickname, bio, avatarURL *string) (*domain.User, error) {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateProfile(user.ID, nickname, bio, avatarURL); err != nil {
		return nil, err
	}

	// Fetch updated user to return
	return s.repo.GetUserByKeycloakID(keycloakID)
}

func (s *UserService) AddAchievement(keycloakID, achievementID string) (*domain.User, error) {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddAchievement(user.ID, achievementID); err != nil {
		return nil, err
	}

	// Fetch updated user to return
	return s.repo.GetUserByKeycloakID(keycloakID)
}

func (s *UserService) GetDailyStats(keycloakID string) (*domain.DailyStats, error) {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetDailyStats(user.ID)
}

func (s *UserService) GetLeaderboard(limit int) ([]domain.LeaderboardUser, error) {
	cacheKey := fmt.Sprintf("leaderboard:%d", limit)
	var cachedResult []domain.LeaderboardUser

	if s.cacheClient != nil {
		if err := s.cacheClient.Get(context.Background(), cacheKey, &cachedResult); err == nil {
			return cachedResult, nil
		}
	}

	result, err := s.repo.GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	if s.cacheClient != nil {
		// Cache the leaderboard for 1 minute
		s.cacheClient.Set(context.Background(), cacheKey, result, 1*time.Minute)
	}

	return result, nil
}

// ==================== Admin Methods ====================

// ListAllUsers returns a paginated list of all users.
func (s *UserService) ListAllUsers(page, limit int, search string) (*domain.UserListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListAllUsers(page, limit, search)
}

// AdminGetUser returns a single user by UUID string.
func (s *UserService) AdminGetUser(userID string) (*domain.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(uid)
}

// BanUser bans a user by UUID string.
func (s *UserService) BanUser(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.repo.BanUser(uid)
}

// UnbanUser unbans a user by UUID string.
func (s *UserService) UnbanUser(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.repo.UnbanUser(uid)
}

// AdminResetProgress resets any user's progress by UUID string.
func (s *UserService) AdminResetProgress(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.repo.AdminResetProgress(uid)
}

// GetPlatformStats returns aggregate platform statistics.
func (s *UserService) GetPlatformStats() (*domain.PlatformStats, error) {
	return s.repo.GetPlatformStats()
}

// GetRegistrationTrend returns registration trend data.
func (s *UserService) GetRegistrationTrend(period string) (*domain.TrendData, error) {
	return s.repo.GetRegistrationTrend(period)
}

// GetCompletionTrend returns lesson completion trend data.
func (s *UserService) GetCompletionTrend(period string) (*domain.TrendData, error) {
	return s.repo.GetCompletionTrend(period)
}

func (s *UserService) RecordLessonTime(keycloakID string, lessonID string, seconds int) error {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return err
	}
	return s.repo.RecordLessonTime(user.ID, lessonID, seconds)
}

func (s *UserService) GetAchievementStats() ([]domain.AchievementStat, error) {
	return s.repo.GetAchievementStats()
}

func (s *UserService) AdminGetUserLessonTimes(userID string) ([]domain.LessonTimeStat, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id format")
	}
	return s.repo.GetUserLessonTimes(uid)
}

func (s *UserService) GetCompletedLessons(keycloakID string) ([]string, error) {
	user, err := s.repo.GetUserByKeycloakID(keycloakID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetCompletedLessons(user.ID)
}

// AdminSetUserRoles updates the roles for a user by UUID string.
func (s *UserService) AdminSetUserRoles(userID string, roleIDs []string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.repo.SetUserRoles(uid, roleIDs)
}

func (s *UserService) GetPublicProfile(userID string) (*domain.PublicProfile, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.GetUserByID(uid)
	if err != nil {
		return nil, err
	}
	return &domain.PublicProfile{
		ID:            user.ID.String(),
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		Bio:           user.Bio,
		XP:            user.XP,
		Points:        user.Points,
		CreatedAt:     user.CreatedAt,
		Achievements:  user.Achievements,
		ActiveEffects: user.ActiveEffects,
	}, nil
}

func (s *UserService) ClaimDailyLogin(userID uuid.UUID) (*domain.DailyLoginResult, error) {
	return s.repo.ClaimDailyLogin(userID)
}

func (s *UserService) GetDailyLoginStatus(userID uuid.UUID, year, month int) (*domain.DailyLoginStatus, error) {
	return s.repo.GetDailyLoginStatus(userID, year, month)
}

func (s *UserService) GetUserInventory(userID uuid.UUID) ([]domain.UserInventoryItem, error) {
	return s.repo.GetUserInventory(userID)
}

func (s *UserService) ToggleEquipItem(userID uuid.UUID, itemID string) (*domain.ToggleEquipResponse, error) {
	if err := s.repo.ToggleEquipItem(userID, itemID); err != nil {
		return nil, err
	}
	items, err := s.repo.GetUserInventory(userID)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return &domain.ToggleEquipResponse{
		Items:         items,
		ActiveEffects: user.ActiveEffects,
	}, nil
}

func (s *UserService) DeleteInventoryItem(userID uuid.UUID, inventoryID string) (*domain.ToggleEquipResponse, error) {
	if err := s.repo.DeleteInventoryItem(userID, inventoryID); err != nil {
		return nil, err
	}
	items, err := s.repo.GetUserInventory(userID)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return &domain.ToggleEquipResponse{
		Items:         items,
		ActiveEffects: user.ActiveEffects,
	}, nil
}

func (s *UserService) GetTodayChallenge(userID uuid.UUID) (*domain.TodayChallengeResponse, error) {
	now := time.Now()
	todayStr := now.Format("2006-01-02")

	challenge, err := s.repo.GetTodayChallenge(todayStr)
	if err != nil {
		return nil, fmt.Errorf("napi kihívás lekérése: %w", err)
	}

	userProgress, err := s.repo.GetUserDailyChallenge(userID, todayStr)
	if err != nil {
		return nil, fmt.Errorf("felhasználói haladás lekérése: %w", err)
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("felhasználó lekérése: %w", err)
	}

	leaderboard, err := s.repo.GetDailyChallengeLeaderboard(todayStr)
	if err != nil {
		leaderboard = []domain.DailyChallengeLeaderboardEntry{}
	}

	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	secondsUntilReset := int(time.Until(tomorrow).Seconds())
	if secondsUntilReset < 0 {
		secondsUntilReset = 0
	}

	resp := &domain.TodayChallengeResponse{
		Challenge:         *challenge,
		CurrentStreak:     user.ChallengeStreak,
		SecondsUntilReset: secondsUntilReset,
		Leaderboard:       leaderboard,
		UserLanguage:      "python",
	}

	if userProgress != nil {
		resp.StartedAt = &userProgress.StartedAt
		resp.CompletedAt = userProgress.CompletedAt
		resp.IsCompleted = userProgress.IsCompleted
		resp.SolveTimeSeconds = userProgress.SolveTimeSeconds
		resp.UserLanguage = userProgress.Language
		if userProgress.SubmittedCode != nil {
			resp.UserSubmittedCode = *userProgress.SubmittedCode
		}
	}

	return resp, nil
}

func (s *UserService) StartDailyChallenge(userID uuid.UUID, language string) (*domain.UserDailyChallenge, error) {
	todayStr := time.Now().Format("2006-01-02")

	challenge, err := s.repo.GetTodayChallenge(todayStr)
	if err != nil {
		return nil, err
	}

	if language == "" {
		language = "python"
	}

	return s.repo.StartDailyChallenge(userID, challenge.ID, todayStr, language)
}

func (s *UserService) RunDailyChallengeCode(language, code string) (*domain.RunChallengeResponse, error) {
	todayStr := time.Now().Format("2006-01-02")
	challenge, err := s.repo.GetTodayChallenge(todayStr)
	if err != nil {
		return nil, err
	}

	return ExecuteChallengeTests(s.pistonClient, challenge.Slug, language, code, challenge.TestCases)
}

func (s *UserService) SubmitDailyChallenge(userID uuid.UUID, language, code string) (*domain.SubmitChallengeResponse, error) {
	todayStr := time.Now().Format("2006-01-02")
	challenge, err := s.repo.GetTodayChallenge(todayStr)
	if err != nil {
		return nil, err
	}

	// 1. Run tests through test runner
	testRunResult, err := ExecuteChallengeTests(s.pistonClient, challenge.Slug, language, code, challenge.TestCases)
	if err != nil {
		return nil, err
	}

	if !testRunResult.Success {
		return &domain.SubmitChallengeResponse{
			Success:       false,
			TestResults:   testRunResult.TestResults,
			ConsoleOutput: testRunResult.ConsoleOutput,
			Error:         "Néhány teszteset elbukott! Javítsd a hibákat a beküldés előtt.",
		}, nil
	}

	// 2. Submit to database and calculate streak / points
	udc, newStreak, err := s.repo.SubmitDailyChallenge(userID, challenge.ID, todayStr, language, code, challenge.PointsReward, challenge.XPReward)
	if err != nil {
		return nil, fmt.Errorf("mentési hiba: %w", err)
	}

	// 3. Get updated leaderboard
	leaderboard, _ := s.repo.GetDailyChallengeLeaderboard(todayStr)
	userRank := 0
	for _, entry := range leaderboard {
		if entry.UserID == userID.String() {
			userRank = entry.Rank
			break
		}
	}

	solveTime := 0
	if udc.SolveTimeSeconds != nil {
		solveTime = *udc.SolveTimeSeconds
	}

	return &domain.SubmitChallengeResponse{
		Success:          true,
		TestResults:      testRunResult.TestResults,
		ConsoleOutput:    testRunResult.ConsoleOutput,
		SolveTimeSeconds: solveTime,
		PointsEarned:     udc.PointsEarned,
		XPEarned:         udc.XPEarned,
		Streak:           newStreak,
		Rank:             userRank,
		Leaderboard:      leaderboard,
	}, nil
}

func (s *UserService) GetDailyChallengeLeaderboard() ([]domain.DailyChallengeLeaderboardEntry, error) {
	todayStr := time.Now().Format("2006-01-02")
	return s.repo.GetDailyChallengeLeaderboard(todayStr)
}




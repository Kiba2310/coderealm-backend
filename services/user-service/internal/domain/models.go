package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AchievementEarned struct {
	ID       string    `json:"id"`
	EarnedAt time.Time `json:"earned_at"`
}

type User struct {
	ID           uuid.UUID           `json:"id"`
	KeycloakID   string              `json:"keycloak_id"`
	Email        string              `json:"email"`
	CreatedAt    time.Time           `json:"created_at"`
	Points       int                 `json:"points"`
	XP           int                 `json:"xp"`
	Nickname     *string             `json:"nickname,omitempty"`
	Bio          *string             `json:"bio,omitempty"`
	AvatarURL         *string             `json:"avatar_url,omitempty"`
	Achievements      []AchievementEarned `json:"achievements"`
	ActiveEffects     []string            `json:"active_effects"`
	Roles             []Role              `json:"roles,omitempty"`
	IsBanned          bool                `json:"is_banned"`
	BannedAt          *time.Time          `json:"banned_at,omitempty"`
	ChallengeStreak   int                 `json:"challenge_streak"`
	LastChallengeDate *string             `json:"last_challenge_date,omitempty"`
}

type AuthResponse struct {
	User *User `json:"user"`
}

type DailyStats struct {
	DailyXP      int `json:"daily_xp"`
	DailyLessons int `json:"daily_lessons"`
}

type LeaderboardUser struct {
	ID        string  `json:"id"`
	Nickname  *string `json:"nickname"`
	AvatarURL    *string  `json:"avatar_url"`
	Points       int      `json:"points"`
	XP           int      `json:"xp"`
	Rank         int      `json:"rank"`
	ActiveEffects []string `json:"active_effects"`
}

type StoreItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Cost          int    `json:"cost"`
	EffectType    string `json:"effect_type"`
	DurationHours int    `json:"duration_hours"`
	Rarity        string `json:"rarity"`
}

type InventoryItem struct {
	ID          string     `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ItemID      string     `json:"item_id"`
	PurchasedAt time.Time  `json:"purchased_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsUsed      bool       `json:"is_used"`
	IsEquipped  bool       `json:"is_equipped"`
}

type UserInventoryItem struct {
	ID            string     `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	ItemID        string     `json:"item_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	EffectType    string     `json:"effect_type"`
	DurationHours int        `json:"duration_hours"`
	Rarity        string     `json:"rarity"`
	PurchasedAt   time.Time  `json:"purchased_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	IsUsed        bool       `json:"is_used"`
	IsEquipped    bool       `json:"is_equipped"`
}

type ToggleEquipResponse struct {
	Items         []UserInventoryItem `json:"items"`
	ActiveEffects []string            `json:"active_effects"`
}

type StoreState struct {
	UserID      uuid.UUID `json:"user_id"`
	Items       []string  `json:"items"`
	LastRefresh time.Time `json:"last_refresh"`
}

// Admin types

type UserListResult struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type PlatformStats struct {
	TotalUsers            int `json:"total_users"`
	ActiveUsersToday      int `json:"active_users_today"`
	TotalLessons          int `json:"total_lessons_completed"`
	TotalXP               int `json:"total_xp_earned"`
	BannedUsers           int `json:"banned_users"`
	TotalTimeSpentSeconds int `json:"total_time_spent_seconds"`
	TotalCodeChallenges   int `json:"total_code_challenges"`
}

type AchievementStat struct {
	AchievementID string  `json:"achievement_id"`
	Percentage    float64 `json:"percentage"`
}

type TrendPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type TrendData struct {
	Points []TrendPoint `json:"points"`
	Period string       `json:"period"`
}

type LessonTimeStat struct {
	LessonID         string `json:"lesson_id"`
	TimeSpentSeconds int    `json:"time_spent_seconds"`
}

type PublicProfile struct {
	ID            string              `json:"id"`
	Nickname      *string             `json:"nickname"`
	AvatarURL     *string             `json:"avatar_url"`
	Bio           *string             `json:"bio"`
	XP            int                 `json:"xp"`
	Points        int                 `json:"points"`
	CreatedAt     time.Time           `json:"created_at"`
	Achievements  []AchievementEarned `json:"achievements"`
	ActiveEffects []string            `json:"active_effects"`
}

type DailyLoginDay struct {
	Date          string `json:"date"`
	StreakDay     int    `json:"streak_day"`
	PointsAwarded int    `json:"points_awarded"`
}

type DailyLoginStatus struct {
	Days           []DailyLoginDay `json:"days"`
	ClaimedToday   bool            `json:"claimed_today"`
	NextStreakDay   int             `json:"next_streak_day"`
	NextPoints     int             `json:"next_points"`
	TotalThisMonth int             `json:"total_this_month"`
}

type DailyLoginResult struct {
	StreakDay      int `json:"streak_day"`
	PointsAwarded int `json:"points_awarded"`
	TotalPoints   int `json:"total_points"`
}

// Daily Challenge types

type DailyChallengeTestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type DailyChallenge struct {
	ID           uuid.UUID                `json:"id"`
	Slug         string                   `json:"slug"`
	Title        string                   `json:"title"`
	Description  string                   `json:"description"`
	Explanation  string                   `json:"explanation"`
	Difficulty   string                   `json:"difficulty"`
	PointsReward int                      `json:"points_reward"`
	XPReward     int                      `json:"xp_reward"`
	StarterCodes map[string]string        `json:"starter_codes"`
	TestCases    []DailyChallengeTestCase `json:"test_cases"`
	CreatedAt    time.Time                `json:"created_at"`
}

type UserDailyChallenge struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	ChallengeID      uuid.UUID  `json:"challenge_id"`
	ChallengeDate    string     `json:"challenge_date"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	SolveTimeSeconds *int       `json:"solve_time_seconds,omitempty"`
	Language         string     `json:"language"`
	SubmittedCode    *string    `json:"submitted_code,omitempty"`
	IsCompleted      bool       `json:"is_completed"`
	PointsEarned     int        `json:"points_earned"`
	XPEarned         int        `json:"xp_earned"`
}

type DailyChallengeLeaderboardEntry struct {
	Rank             int       `json:"rank"`
	UserID           string    `json:"user_id"`
	Nickname         string    `json:"nickname"`
	AvatarURL        *string   `json:"avatar_url"`
	ActiveEffects    []string  `json:"active_effects"`
	Language         string    `json:"language"`
	SolveTimeSeconds int       `json:"solve_time_seconds"`
	CompletedAt      time.Time `json:"completed_at"`
	BonusPoints      int       `json:"bonus_points"`
}

type TodayChallengeResponse struct {
	Challenge         DailyChallenge                   `json:"challenge"`
	StartedAt         *time.Time                       `json:"started_at,omitempty"`
	CompletedAt       *time.Time                       `json:"completed_at,omitempty"`
	IsCompleted       bool                             `json:"is_completed"`
	SolveTimeSeconds  *int                             `json:"solve_time_seconds,omitempty"`
	CurrentStreak     int                              `json:"current_streak"`
	UserLanguage      string                           `json:"user_language"`
	UserSubmittedCode string                           `json:"user_submitted_code,omitempty"`
	SecondsUntilReset int                              `json:"seconds_until_reset"`
	Leaderboard       []DailyChallengeLeaderboardEntry `json:"leaderboard"`
}

type TestResultItem struct {
	Index    int    `json:"index"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
}

type RunChallengeResponse struct {
	Success       bool             `json:"success"`
	TestResults   []TestResultItem `json:"test_results"`
	ConsoleOutput string           `json:"console_output"`
	TotalPassed   int              `json:"total_passed"`
	TotalTests    int              `json:"total_tests"`
	Error         string           `json:"error,omitempty"`
}

type SubmitChallengeResponse struct {
	Success          bool                             `json:"success"`
	TestResults      []TestResultItem                 `json:"test_results"`
	ConsoleOutput    string                           `json:"console_output"`
	SolveTimeSeconds int                              `json:"solve_time_seconds"`
	PointsEarned     int                              `json:"points_earned"`
	XPEarned         int                              `json:"xp_earned"`
	Streak           int                              `json:"streak"`
	Rank             int                              `json:"rank"`
	Leaderboard      []DailyChallengeLeaderboardEntry `json:"leaderboard"`
	Error            string                           `json:"error,omitempty"`
}


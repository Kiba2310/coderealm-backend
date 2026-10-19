package repository

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"BootDevClone/services/user-service/internal/domain"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes all SQL files in the migrations directory in order.
func RunMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("run migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// scanUser is a helper that scans a user row including ban fields and challenge streak.
func (r *UserRepository) scanUser(row interface{ Scan(dest ...any) error }) (*domain.User, error) {
	u := &domain.User{}
	var achievementsJSON []byte
	var activeEffectsJSON []byte
	var bannedAt sql.NullTime
	var lastChallengeDate sql.NullString
	err := row.Scan(&u.ID, &u.KeycloakID, &u.Email, &u.CreatedAt, &u.Points, &u.XP,
		&u.Nickname, &u.Bio, &u.AvatarURL, &achievementsJSON, &activeEffectsJSON, &u.IsBanned, &bannedAt,
		&u.ChallengeStreak, &lastChallengeDate)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if achievementsJSON != nil {
		if err := json.Unmarshal(achievementsJSON, &u.Achievements); err != nil {
			return nil, err
		}
	}
	if u.Achievements == nil {
		u.Achievements = []domain.AchievementEarned{}
	}
	if activeEffectsJSON != nil {
		if err := json.Unmarshal(activeEffectsJSON, &u.ActiveEffects); err != nil {
			return nil, err
		}
	}
	if u.ActiveEffects == nil {
		u.ActiveEffects = []string{}
	}
	if bannedAt.Valid {
		u.BannedAt = &bannedAt.Time
	}
	if lastChallengeDate.Valid {
		u.LastChallengeDate = &lastChallengeDate.String
	}

	roles, err := r.GetUserRoles(u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return u, nil
}

const userSelectCols = `id, keycloak_id, email, created_at, points, xp, nickname, bio, avatar_url, 
(SELECT COALESCE(json_agg(json_build_object('id', achievement_id, 'earned_at', earned_at) ORDER BY earned_at ASC), '[]'::json) FROM user_achievements WHERE user_id = users.id) as achievements, 
(SELECT COALESCE(json_agg(s.id), '[]'::json) FROM user_inventory i JOIN store_items s ON i.item_id = s.id WHERE i.user_id = users.id AND i.is_used = false AND COALESCE(i.is_equipped, true) = true AND (i.expires_at IS NULL OR i.expires_at > NOW())) as active_effects,
COALESCE(is_banned, false), banned_at, COALESCE(challenge_streak, 0), last_challenge_date::text`

func (r *UserRepository) CreateUser(u *domain.User) error {
	_, err := r.db.Exec(`
		INSERT INTO users (id, keycloak_id, email, nickname, bio, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.KeycloakID, u.Email, u.Nickname, u.Bio, u.AvatarURL,
	)
	return err
}

func (r *UserRepository) GetUserByEmail(email string) (*domain.User, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM users WHERE email = $1`, userSelectCols), email)
	return r.scanUser(row)
}

func (r *UserRepository) GetUserByKeycloakID(keycloakID string) (*domain.User, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM users WHERE keycloak_id = $1`, userSelectCols), keycloakID)
	return r.scanUser(row)
}

func (r *UserRepository) AssignRole(userID uuid.UUID, roleID string) error {
	_, err := r.db.Exec(`
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
		userID, roleID,
	)
	return err
}

func (r *UserRepository) RemoveRole(userID uuid.UUID, roleID string) error {
	_, err := r.db.Exec(`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`, userID, roleID)
	return err
}

// SetUserRoles replaces all roles for a user with the given set of role IDs.
func (r *UserRepository) SetUserRoles(userID uuid.UUID, roleIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing roles
	if _, err := tx.Exec(`DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}

	// Insert new roles
	for _, roleID := range roleIDs {
		if _, err := tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *UserRepository) GetUserRoles(userID uuid.UUID) ([]domain.Role, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.name 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userSelectCols), id)
	return r.scanUser(row)
}

// CompleteLessonAndAwardPoints atomically records a lesson completion and awards points.
// Returns (true, newTotalPoints, nil) if points were awarded, (false, currentPoints, nil) if already completed.
func (r *UserRepository) CompleteLessonAndAwardPoints(userID uuid.UUID, lessonID string, lessonType string, points int) (bool, int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO completed_lessons (user_id, lesson_id, lesson_type, points_earned)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, lesson_id) DO NOTHING`,
		userID, lessonID, lessonType, points,
	)
	if err != nil {
		return false, 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, 0, err
	}

	if rowsAffected == 0 {
		var currentPoints int
		err = tx.QueryRow(`SELECT points FROM users WHERE id = $1`, userID).Scan(&currentPoints)
		if err != nil {
			return false, 0, err
		}
		return false, currentPoints, tx.Commit()
	}

	var newPoints int
	err = tx.QueryRow(`
		UPDATE users SET points = points + $1, xp = xp + $1 WHERE id = $2 RETURNING points`,
		points, userID,
	).Scan(&newPoints)
	if err != nil {
		return false, 0, err
	}

	return true, newPoints, tx.Commit()
}

// ResetProgress deletes all completed lessons for a user and sets points to 0.
func (r *UserRepository) ResetProgress(userID uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM completed_lessons WHERE user_id = $1`, userID); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE users SET points = 0, xp = 0 WHERE id = $1`, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) UpdateProfile(userID uuid.UUID, nickname, bio, avatarURL *string) error {
	_, err := r.db.Exec(`
		UPDATE users 
		SET nickname = COALESCE($1, nickname),
		    bio = COALESCE($2, bio),
		    avatar_url = COALESCE($3, avatar_url)
		WHERE id = $4`,
		nickname, bio, avatarURL, userID,
	)
	return err
}

// AddAchievement adds an achievement to the user_achievements table.
func (r *UserRepository) AddAchievement(userID uuid.UUID, achievementID string) error {
	_, err := r.db.Exec(`
		INSERT INTO user_achievements (user_id, achievement_id, earned_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, achievement_id) DO NOTHING`,
		userID, achievementID,
	)
	return err
}

func (r *UserRepository) GetDailyStats(userID uuid.UUID) (*domain.DailyStats, error) {
	var stats domain.DailyStats
	err := r.db.QueryRow(`
		SELECT 
			COALESCE(SUM(points_earned), 0) as daily_xp,
			COUNT(lesson_id) as daily_lessons
		FROM completed_lessons
		WHERE user_id = $1 AND completed_at >= CURRENT_DATE`,
		userID,
	).Scan(&stats.DailyXP, &stats.DailyLessons)

	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *UserRepository) GetLeaderboard(limit int) ([]domain.LeaderboardUser, error) {
	rows, err := r.db.Query(`
		SELECT id, COALESCE(NULLIF(nickname, ''), split_part(email, '@', 1)), avatar_url, points, xp, 
		       RANK() OVER (ORDER BY xp DESC) as rank,
			   (SELECT COALESCE(json_agg(s.id), '[]'::json) FROM user_inventory i JOIN store_items s ON i.item_id = s.id WHERE i.user_id = users.id AND i.is_used = false AND COALESCE(i.is_equipped, true) = true AND (i.expires_at IS NULL OR i.expires_at > NOW())) as active_effects
		FROM users
		WHERE xp > 0
		ORDER BY xp DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.LeaderboardUser
	for rows.Next() {
		var u domain.LeaderboardUser
		var activeEffectsJSON []byte
		if err := rows.Scan(&u.ID, &u.Nickname, &u.AvatarURL, &u.Points, &u.XP, &u.Rank, &activeEffectsJSON); err != nil {
			return nil, err
		}
		if activeEffectsJSON != nil {
			json.Unmarshal(activeEffectsJSON, &u.ActiveEffects)
		}
		if u.ActiveEffects == nil {
			u.ActiveEffects = []string{}
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// ==================== Admin Methods ====================

// ListAllUsers returns a paginated list of users with optional search.
func (r *UserRepository) ListAllUsers(page, limit int, search string) (*domain.UserListResult, error) {
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM users`
	var totalCount int
	if search != "" {
		countQuery = `SELECT COUNT(*) FROM users WHERE email ILIKE $1 OR nickname ILIKE $1`
		err := r.db.QueryRow(countQuery, "%"+search+"%").Scan(&totalCount)
		if err != nil {
			return nil, fmt.Errorf("count users: %w", err)
		}
	} else {
		err := r.db.QueryRow(countQuery).Scan(&totalCount)
		if err != nil {
			return nil, fmt.Errorf("count users: %w", err)
		}
	}

	var listQuery string
	var rows *sql.Rows
	var err error
	if search != "" {
		listQuery = fmt.Sprintf(`SELECT %s FROM users WHERE email ILIKE $1 OR nickname ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userSelectCols)
		rows, err = r.db.Query(listQuery, "%"+search+"%", limit, offset)
	} else {
		listQuery = fmt.Sprintf(`SELECT %s FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, userSelectCols)
		rows, err = r.db.Query(listQuery, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, *u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if users == nil {
		users = []domain.User{}
	}

	return &domain.UserListResult{
		Users:      users,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}, nil
}

// TouchUserActivity updates the last_active_at timestamp for a user.
func (r *UserRepository) TouchUserActivity(userID uuid.UUID) {
	_, _ = r.db.Exec(`UPDATE users SET last_active_at = NOW() WHERE id = $1`, userID)
}

// BanUser sets is_banned=true and records banned_at timestamp.
func (r *UserRepository) BanUser(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE users SET is_banned = true, banned_at = $1 WHERE id = $2`,
		time.Now(), userID)
	return err
}

// UnbanUser clears the ban flag.
func (r *UserRepository) UnbanUser(userID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE users SET is_banned = false, banned_at = NULL WHERE id = $1`, userID)
	return err
}

// AdminResetProgress resets progress for any user by UUID.
func (r *UserRepository) AdminResetProgress(userID uuid.UUID) error {
	return r.ResetProgress(userID)
}

// GetPlatformStats returns aggregate platform statistics.
func (r *UserRepository) GetPlatformStats() (*domain.PlatformStats, error) {
	stats := &domain.PlatformStats{}
	r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	r.db.QueryRow(`
		SELECT COUNT(DISTINCT id) FROM users 
		WHERE (last_active_at IS NOT NULL AND last_active_at >= CURRENT_DATE)
		   OR created_at >= CURRENT_DATE 
		   OR id IN (SELECT user_id FROM completed_lessons WHERE completed_at >= CURRENT_DATE)
		   OR id IN (SELECT user_id FROM daily_logins WHERE login_date = CURRENT_DATE)
		   OR id IN (SELECT user_id FROM user_daily_challenges WHERE started_at >= CURRENT_DATE)
	`).Scan(&stats.ActiveUsersToday)
	r.db.QueryRow(`SELECT COUNT(*) FROM completed_lessons`).Scan(&stats.TotalLessons)
	r.db.QueryRow(`SELECT COALESCE(SUM(points), 0) FROM users`).Scan(&stats.TotalXP)
	r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_banned = true`).Scan(&stats.BannedUsers)
	r.db.QueryRow(`SELECT COALESCE(SUM(time_spent_seconds), 0) FROM lesson_time_tracking`).Scan(&stats.TotalTimeSpentSeconds)
	r.db.QueryRow(`SELECT COUNT(*) FROM completed_lessons WHERE lesson_type = 'code_challenge'`).Scan(&stats.TotalCodeChallenges)
	return stats, nil
}

// GetRegistrationTrend returns registration counts grouped by period.
func (r *UserRepository) GetRegistrationTrend(period string) (*domain.TrendData, error) {
	var query string
	switch period {
	case "daily":
		query = `SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY date ORDER BY date`
	case "weekly":
		query = `SELECT TO_CHAR(DATE_TRUNC('week', created_at), 'YYYY-MM-DD') as date, COUNT(*) as count FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '12 weeks' GROUP BY date ORDER BY date`
	case "monthly":
		query = `SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') as date, COUNT(*) as count FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '12 months' GROUP BY date ORDER BY date`
	case "yearly":
		query = `SELECT TO_CHAR(DATE_TRUNC('year', created_at), 'YYYY') as date, COUNT(*) as count FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '5 years' GROUP BY date ORDER BY date`
	default:
		query = `SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count FROM users WHERE created_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY date ORDER BY date`
		period = "daily"
	}
	return r.queryTrend(query, period)
}

// GetCompletionTrend returns lesson completion counts grouped by period.
func (r *UserRepository) GetCompletionTrend(period string) (*domain.TrendData, error) {
	var query string
	switch period {
	case "daily":
		query = `SELECT TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COUNT(*) as count FROM completed_lessons WHERE completed_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY date ORDER BY date`
	case "weekly":
		query = `SELECT TO_CHAR(DATE_TRUNC('week', completed_at), 'YYYY-MM-DD') as date, COUNT(*) as count FROM completed_lessons WHERE completed_at >= CURRENT_DATE - INTERVAL '12 weeks' GROUP BY date ORDER BY date`
	case "monthly":
		query = `SELECT TO_CHAR(DATE_TRUNC('month', completed_at), 'YYYY-MM') as date, COUNT(*) as count FROM completed_lessons WHERE completed_at >= CURRENT_DATE - INTERVAL '12 months' GROUP BY date ORDER BY date`
	case "yearly":
		query = `SELECT TO_CHAR(DATE_TRUNC('year', completed_at), 'YYYY') as date, COUNT(*) as count FROM completed_lessons WHERE completed_at >= CURRENT_DATE - INTERVAL '5 years' GROUP BY date ORDER BY date`
	default:
		query = `SELECT TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COUNT(*) as count FROM completed_lessons WHERE completed_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY date ORDER BY date`
		period = "daily"
	}
	return r.queryTrend(query, period)
}

func (r *UserRepository) queryTrend(query, period string) (*domain.TrendData, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query trend: %w", err)
	}
	defer rows.Close()

	var points []domain.TrendPoint
	for rows.Next() {
		var p domain.TrendPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, fmt.Errorf("scan trend point: %w", err)
		}
		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if points == nil {
		points = []domain.TrendPoint{}
	}

	return &domain.TrendData{
		Points: points,
		Period: period,
	}, nil
}

// RecordLessonTime records or updates the time spent on a lesson by a user
func (r *UserRepository) RecordLessonTime(userID uuid.UUID, lessonID string, seconds int) error {
	_, err := r.db.Exec(`
		INSERT INTO lesson_time_tracking (user_id, lesson_id, time_spent_seconds)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, lesson_id)
		DO UPDATE SET time_spent_seconds = lesson_time_tracking.time_spent_seconds + EXCLUDED.time_spent_seconds
	`, userID, lessonID, seconds)
	return err
}

// GetUserTotalCompletedLessonsCount returns the total number of distinct completed lessons for a user.
func (r *UserRepository) GetUserTotalCompletedLessonsCount(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(DISTINCT lesson_id) FROM completed_lessons WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

// GetUserCodeChallengeCount returns the total number of distinct completed code challenges for a user.
func (r *UserRepository) GetUserCodeChallengeCount(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT lesson_id) 
		FROM completed_lessons 
		WHERE user_id = $1 AND lesson_type = 'code_challenge'
	`, userID).Scan(&count)
	return count, err
}

// GetUserLessonTimes retrieves all lesson time stats for a given user
func (r *UserRepository) GetUserLessonTimes(userID uuid.UUID) ([]domain.LessonTimeStat, error) {
	rows, err := r.db.Query(`
		SELECT lesson_id, time_spent_seconds
		FROM lesson_time_tracking
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query lesson times: %w", err)
	}
	defer rows.Close()

	var stats []domain.LessonTimeStat
	for rows.Next() {
		var stat domain.LessonTimeStat
		if err := rows.Scan(&stat.LessonID, &stat.TimeSpentSeconds); err != nil {
			return nil, fmt.Errorf("scan lesson time: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if stats == nil {
		stats = []domain.LessonTimeStat{}
	}
	return stats, nil
}

// GetCompletedLessons returns a list of lesson IDs completed by the user
func (r *UserRepository) GetCompletedLessons(userID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT lesson_id
		FROM completed_lessons
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query completed lessons: %w", err)
	}
	defer rows.Close()

	var lessons []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan completed lesson: %w", err)
		}
		lessons = append(lessons, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if lessons == nil {
		lessons = []string{}
	}
	return lessons, nil
}

func (r *UserRepository) ClaimDailyLogin(userID uuid.UUID) (*domain.DailyLoginResult, error) {
	today := time.Now().Format("2006-01-02")
	currentMonth := time.Now().Format("2006-01")

	// Check if already claimed today
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM daily_logins WHERE user_id = $1 AND login_date = $2)`, userID, today).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("ma már beváltottad a napi jutalmat")
	}

	// Count logins this month to determine streak_day
	var countThisMonth int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM daily_logins WHERE user_id = $1 AND login_date >= $2::date`, userID, currentMonth+"-01").Scan(&countThisMonth)
	if err != nil {
		return nil, err
	}

	streakDay := countThisMonth + 1
	pointsToAward := streakDay * 5

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO daily_logins (user_id, login_date, streak_day, points_awarded) VALUES ($1, $2, $3, $4)`,
		userID, today, streakDay, pointsToAward)
	if err != nil {
		return nil, err
	}

	var totalPoints int
	err = tx.QueryRow(`UPDATE users SET points = points + $1 WHERE id = $2 RETURNING points`, pointsToAward, userID).Scan(&totalPoints)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &domain.DailyLoginResult{
		StreakDay:      streakDay,
		PointsAwarded: pointsToAward,
		TotalPoints:   totalPoints,
	}, nil
}

func (r *UserRepository) GetDailyLoginStatus(userID uuid.UUID, year, month int) (*domain.DailyLoginStatus, error) {
	monthStr := fmt.Sprintf("%d-%02d-01", year, month)
	nextMonthStr := fmt.Sprintf("%d-%02d-01", year, month+1)
	if month == 12 {
		nextMonthStr = fmt.Sprintf("%d-01-01", year+1)
	}

	rows, err := r.db.Query(`
		SELECT login_date, streak_day, points_awarded 
		FROM daily_logins 
		WHERE user_id = $1 AND login_date >= $2::date AND login_date < $3::date
		ORDER BY login_date ASC`, userID, monthStr, nextMonthStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []domain.DailyLoginDay
	claimedToday := false
	today := time.Now().Format("2006-01-02")
	totalPoints := 0

	for rows.Next() {
		var d domain.DailyLoginDay
		var loginDate time.Time
		if err := rows.Scan(&loginDate, &d.StreakDay, &d.PointsAwarded); err != nil {
			return nil, err
		}
		d.Date = loginDate.Format("2006-01-02")
		if d.Date == today {
			claimedToday = true
		}
		totalPoints += d.PointsAwarded
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if days == nil {
		days = []domain.DailyLoginDay{}
	}

	nextStreakDay := len(days) + 1
	if claimedToday {
		nextStreakDay = len(days) + 1 // Already includes today
	}
	nextPoints := nextStreakDay * 5
	if claimedToday {
		nextStreakDay = 0
		nextPoints = 0
	}

	return &domain.DailyLoginStatus{
		Days:           days,
		ClaimedToday:   claimedToday,
		NextStreakDay:   nextStreakDay,
		NextPoints:     nextPoints,
		TotalThisMonth: totalPoints,
	}, nil
}

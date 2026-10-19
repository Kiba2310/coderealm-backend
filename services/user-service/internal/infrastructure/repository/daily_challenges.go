package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"BootDevClone/services/user-service/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (r *UserRepository) GetAllDailyChallenges() ([]domain.DailyChallenge, error) {
	rows, err := r.db.Query(`
		SELECT id, slug, title, description, explanation, difficulty, points_reward, xp_reward, starter_codes, test_cases, created_at
		FROM daily_challenges
		ORDER BY slug ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query daily challenges: %w", err)
	}
	defer rows.Close()

	var challenges []domain.DailyChallenge
	for rows.Next() {
		var c domain.DailyChallenge
		var starterCodesJSON, testCasesJSON []byte

		if err := rows.Scan(
			&c.ID, &c.Slug, &c.Title, &c.Description, &c.Explanation,
			&c.Difficulty, &c.PointsReward, &c.XPReward,
			&starterCodesJSON, &testCasesJSON, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan daily challenge: %w", err)
		}

		if err := json.Unmarshal(starterCodesJSON, &c.StarterCodes); err != nil {
			c.StarterCodes = make(map[string]string)
		}
		if err := json.Unmarshal(testCasesJSON, &c.TestCases); err != nil {
			c.TestCases = []domain.DailyChallengeTestCase{}
		}

		challenges = append(challenges, c)
	}

	return challenges, nil
}

func (r *UserRepository) GetTodayChallenge(dateStr string) (*domain.DailyChallenge, error) {
	challenges, err := r.GetAllDailyChallenges()
	if err != nil {
		return nil, err
	}
	if len(challenges) == 0 {
		return nil, fmt.Errorf("nincsenek elérhető napi kihívások az adatbázisban")
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t = time.Now()
	}

	// Deterministic index calculation based on day of year
	dayOfYear := t.YearDay()
	index := (dayOfYear + t.Year()) % len(challenges)

	challenge := challenges[index]
	return &challenge, nil
}

func (r *UserRepository) GetUserDailyChallenge(userID uuid.UUID, challengeDate string) (*domain.UserDailyChallenge, error) {
	var udc domain.UserDailyChallenge
	var solveTime sql.NullInt64
	var submittedCode sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, user_id, challenge_id, challenge_date::text, started_at, completed_at, solve_time_seconds, language, submitted_code, is_completed, points_earned, xp_earned
		FROM user_daily_challenges
		WHERE user_id = $1 AND challenge_date = $2
	`, userID, challengeDate).Scan(
		&udc.ID, &udc.UserID, &udc.ChallengeID, &udc.ChallengeDate,
		&udc.StartedAt, &completedAt, &solveTime, &udc.Language,
		&submittedCode, &udc.IsCompleted, &udc.PointsEarned, &udc.XPEarned,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if completedAt.Valid {
		udc.CompletedAt = &completedAt.Time
	}
	if solveTime.Valid {
		st := int(solveTime.Int64)
		udc.SolveTimeSeconds = &st
	}
	if submittedCode.Valid {
		udc.SubmittedCode = &submittedCode.String
	}

	return &udc, nil
}

func (r *UserRepository) StartDailyChallenge(userID uuid.UUID, challengeID uuid.UUID, challengeDate string, language string) (*domain.UserDailyChallenge, error) {
	existing, err := r.GetUserDailyChallenge(userID, challengeDate)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	var udc domain.UserDailyChallenge
	err = r.db.QueryRow(`
		INSERT INTO user_daily_challenges (user_id, challenge_id, challenge_date, language, started_at, is_completed)
		VALUES ($1, $2, $3, $4, NOW(), false)
		ON CONFLICT (user_id, challenge_date) DO UPDATE SET language = EXCLUDED.language
		RETURNING id, user_id, challenge_id, challenge_date::text, started_at, language, is_completed
	`, userID, challengeID, challengeDate, language).Scan(
		&udc.ID, &udc.UserID, &udc.ChallengeID, &udc.ChallengeDate,
		&udc.StartedAt, &udc.Language, &udc.IsCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("start daily challenge: %w", err)
	}

	return &udc, nil
}

func (r *UserRepository) SubmitDailyChallenge(userID uuid.UUID, challengeID uuid.UUID, challengeDate string, language, submittedCode string, points, xp int) (*domain.UserDailyChallenge, int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	// Ensure started_at exists
	var startedAt time.Time
	err = tx.QueryRow(`
		SELECT started_at FROM user_daily_challenges
		WHERE user_id = $1 AND challenge_date = $2
	`, userID, challengeDate).Scan(&startedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			startedAt = time.Now()
			_, _ = tx.Exec(`
				INSERT INTO user_daily_challenges (user_id, challenge_id, challenge_date, language, started_at)
				VALUES ($1, $2, $3, $4, $5)
			`, userID, challengeID, challengeDate, language, startedAt)
		} else {
			return nil, 0, err
		}
	}

	now := time.Now()
	solveTimeSeconds := int(now.Sub(startedAt).Seconds())
	if solveTimeSeconds < 1 {
		solveTimeSeconds = 1
	}

	var udc domain.UserDailyChallenge
	var completedAt sql.NullTime
	var solveTime sql.NullInt64

	err = tx.QueryRow(`
		UPDATE user_daily_challenges
		SET is_completed = true, completed_at = NOW(), solve_time_seconds = $3,
		    language = $4, submitted_code = $5, points_earned = $6, xp_earned = $7
		WHERE user_id = $1 AND challenge_date = $2
		RETURNING id, user_id, challenge_id, challenge_date::text, started_at, completed_at, solve_time_seconds, language, submitted_code, is_completed, points_earned, xp_earned
	`, userID, challengeDate, solveTimeSeconds, language, submittedCode, points, xp).Scan(
		&udc.ID, &udc.UserID, &udc.ChallengeID, &udc.ChallengeDate,
		&udc.StartedAt, &completedAt, &solveTime, &udc.Language,
		&submittedCode, &udc.IsCompleted, &udc.PointsEarned, &udc.XPEarned,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("update user daily challenge: %w", err)
	}

	if completedAt.Valid {
		udc.CompletedAt = &completedAt.Time
	}
	if solveTime.Valid {
		st := int(solveTime.Int64)
		udc.SolveTimeSeconds = &st
	}
	udc.SubmittedCode = &submittedCode

	// Update user streak and points
	var currentStreak int
	var lastChallengeDate sql.NullTime
	err = tx.QueryRow(`SELECT challenge_streak, last_challenge_date FROM users WHERE id = $1`, userID).Scan(&currentStreak, &lastChallengeDate)
	if err != nil {
		return nil, 0, err
	}

	today, _ := time.Parse("2006-01-02", challengeDate)
	yesterday := today.AddDate(0, 0, -1)

	newStreak := currentStreak
	if !lastChallengeDate.Valid {
		newStreak = 1
	} else {
		lastDate := lastChallengeDate.Time.Format("2006-01-02")
		if lastDate == yesterday.Format("2006-01-02") {
			newStreak = currentStreak + 1
		} else if lastDate == today.Format("2006-01-02") {
			// Already completed today, keep streak
			newStreak = currentStreak
		} else {
			// Streak broken
			newStreak = 1
		}
	}

	_, err = tx.Exec(`
		UPDATE users
		SET points = points + $1, xp = xp + $2, challenge_streak = $3, last_challenge_date = $4
		WHERE id = $5
	`, points, xp, newStreak, challengeDate, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("update user points and streak: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	return &udc, newStreak, nil
}

func (r *UserRepository) GetDailyChallengeLeaderboard(challengeDate string) ([]domain.DailyChallengeLeaderboardEntry, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.nickname, u.email, u.avatar_url,
		       COALESCE(udc.language, 'python') as language,
		       COALESCE(udc.solve_time_seconds, 0) as solve_time_seconds,
		       udc.completed_at,
		       COALESCE((
		           SELECT array_agg(DISTINCT i.item_id)
		           FROM user_inventory i
		           WHERE i.user_id = u.id AND i.is_used = false AND COALESCE(i.is_equipped, true) = true
		       ), '{}') as active_effects
		FROM user_daily_challenges udc
		JOIN users u ON udc.user_id = u.id
		WHERE udc.challenge_date = $1 AND udc.is_completed = true
		ORDER BY udc.solve_time_seconds ASC, udc.completed_at ASC
		LIMIT 50
	`, challengeDate)
	if err != nil {
		return nil, fmt.Errorf("query daily leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []domain.DailyChallengeLeaderboardEntry
	rank := 1

	for rows.Next() {
		var entry domain.DailyChallengeLeaderboardEntry
		var email string
		var rawAvatar *string
		var effects pq.StringArray

		if err := rows.Scan(
			&entry.UserID, &entry.Nickname, &email, &rawAvatar,
			&entry.Language, &entry.SolveTimeSeconds, &entry.CompletedAt, &effects,
		); err != nil {
			return nil, fmt.Errorf("scan daily leaderboard entry: %w", err)
		}

		if entry.Nickname == "" {
			entry.Nickname = email
		}
		entry.AvatarURL = rawAvatar
		entry.ActiveEffects = []string(effects)
		entry.Rank = rank

		// Bonus rewards for Top 3
		switch rank {
		case 1:
			entry.BonusPoints = 100
		case 2:
			entry.BonusPoints = 50
		case 3:
			entry.BonusPoints = 25
		default:
			entry.BonusPoints = 0
		}

		entries = append(entries, entry)
		rank++
	}

	if entries == nil {
		entries = []domain.DailyChallengeLeaderboardEntry{}
	}

	return entries, nil
}

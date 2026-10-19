package repository

import "BootDevClone/services/user-service/internal/domain"

// GetAchievementStats returns the global completion percentage for each achievement.
func (r *UserRepository) GetAchievementStats() ([]domain.AchievementStat, error) {
	rows, err := r.db.Query(`
		WITH UserCount AS (
			SELECT NULLIF(COUNT(*), 0) as total FROM users
		)
		SELECT ua.achievement_id, 
		       COUNT(ua.user_id)::float * 100.0 / uc.total as percentage
		FROM user_achievements ua
		CROSS JOIN UserCount uc
		GROUP BY ua.achievement_id, uc.total
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.AchievementStat
	for rows.Next() {
		var stat domain.AchievementStat
		if err := rows.Scan(&stat.AchievementID, &stat.Percentage); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if stats == nil {
		stats = []domain.AchievementStat{}
	}
	return stats, nil
}

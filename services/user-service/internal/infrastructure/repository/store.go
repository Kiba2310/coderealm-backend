package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"BootDevClone/services/user-service/internal/domain"
	"github.com/google/uuid"
)

func (r *UserRepository) GetStoreItems(userID uuid.UUID) ([]domain.StoreItem, error) {
	// First, try to get user's current store state
	var stateJSON []byte
	var lastRefresh time.Time
	err := r.db.QueryRow(`SELECT items, last_refresh FROM user_store_state WHERE user_id = $1`, userID).Scan(&stateJSON, &lastRefresh)

	// If no state or last refresh was before midnight today, refresh
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if err == sql.ErrNoRows || lastRefresh.Before(midnight) {
		return r.RefreshStore(userID, 0) // free refresh
	} else if err != nil {
		return nil, err
	}

	var itemIDs []string
	if err := json.Unmarshal(stateJSON, &itemIDs); err != nil {
		return nil, err
	}

	return r.getItemsByIDs(itemIDs)
}

func (r *UserRepository) RefreshStore(userID uuid.UUID, cost int) ([]domain.StoreItem, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if cost > 0 {
		res, err := tx.Exec(`UPDATE users SET points = points - $1 WHERE id = $2 AND points >= $1`, cost, userID)
		if err != nil {
			return nil, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil, sql.ErrNoRows // Not enough points
		}
	}

	rows, err := tx.Query(`SELECT id FROM store_items`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allItemIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		allItemIDs = append(allItemIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Shuffle and pick 3
	rand.Shuffle(len(allItemIDs), func(i, j int) {
		allItemIDs[i], allItemIDs[j] = allItemIDs[j], allItemIDs[i]
	})
	
	selectedIDs := allItemIDs
	if len(selectedIDs) > 3 {
		selectedIDs = selectedIDs[:3]
	}

	selectedJSON, _ := json.Marshal(selectedIDs)

	_, err = tx.Exec(`
		INSERT INTO user_store_state (user_id, items, last_refresh) 
		VALUES ($1, $2, NOW()) 
		ON CONFLICT (user_id) DO UPDATE SET items = EXCLUDED.items, last_refresh = EXCLUDED.last_refresh`,
		userID, selectedJSON)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.getItemsByIDs(selectedIDs)
}

func (r *UserRepository) getItemsByIDs(itemIDs []string) ([]domain.StoreItem, error) {
	if len(itemIDs) == 0 {
		return []domain.StoreItem{}, nil
	}

	rows, err := r.db.Query(`SELECT id, name, description, cost, effect_type, duration_hours, rarity FROM store_items`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allItems []domain.StoreItem
	for rows.Next() {
		var item domain.StoreItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Cost, &item.EffectType, &item.DurationHours, &item.Rarity); err != nil {
			return nil, err
		}
		allItems = append(allItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var result []domain.StoreItem
	for _, id := range itemIDs {
		for _, item := range allItems {
			if item.ID == id {
				result = append(result, item)
				break
			}
		}
	}

	return result, nil
}

func (r *UserRepository) PurchaseItem(userID uuid.UUID, itemID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var cost, duration int
	var effectType string
	err = tx.QueryRow(`SELECT cost, duration_hours, effect_type FROM store_items WHERE id = $1`, itemID).Scan(&cost, &duration, &effectType)
	if err != nil {
		return err
	}

	res, err := tx.Exec(`UPDATE users SET points = points - $1 WHERE id = $2 AND points >= $1`, cost, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows // Not enough points
	}

	var expiresAt *time.Time
	if duration > 0 {
		exp := time.Now().Add(time.Duration(duration) * time.Hour)
		expiresAt = &exp
	}

	_, err = tx.Exec(`
		INSERT INTO user_inventory (user_id, item_id, expires_at)
		VALUES ($1, $2, $3)`,
		userID, itemID, expiresAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) ConsumeHelperItem(userID uuid.UUID, lessonID string) (bool, error) {
	res, err := r.db.Exec(`
		UPDATE user_inventory 
		SET is_used = true, used_on_lesson_id = $2
		WHERE id = (
			SELECT i.id FROM user_inventory i 
			JOIN store_items s ON i.item_id = s.id 
			WHERE i.user_id = $1 AND s.effect_type = 'free_help' AND i.is_used = false 
			LIMIT 1
		)`, userID, lessonID)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (r *UserRepository) CheckIfHelperUsedOnLesson(userID uuid.UUID, lessonID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM user_inventory 
		WHERE user_id = $1 AND used_on_lesson_id = $2 AND is_used = true
	`, userID, lessonID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) HasActiveXPAmulet(userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM user_inventory i 
			JOIN store_items s ON i.item_id = s.id 
			WHERE i.user_id = $1 AND s.effect_type = 'xp_boost' AND (i.expires_at IS NULL OR i.expires_at > NOW())
		)`, userID).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetUserInventory(userID uuid.UUID) ([]domain.UserInventoryItem, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.user_id, i.item_id, s.name, s.description, s.effect_type, s.duration_hours, s.rarity,
		       i.purchased_at, i.expires_at, i.is_used, COALESCE(i.is_equipped, true)
		FROM user_inventory i
		JOIN store_items s ON i.item_id = s.id
		WHERE i.user_id = $1
		ORDER BY i.purchased_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.UserInventoryItem
	for rows.Next() {
		var item domain.UserInventoryItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.ItemID, &item.Name, &item.Description, &item.EffectType, &item.DurationHours, &item.Rarity,
			&item.PurchasedAt, &item.ExpiresAt, &item.IsUsed, &item.IsEquipped,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.UserInventoryItem{}
	}
	return items, nil
}

func (r *UserRepository) ToggleEquipItem(userID uuid.UUID, itemID string) error {
	res, err := r.db.Exec(`
		UPDATE user_inventory
		SET is_equipped = NOT COALESCE(is_equipped, true)
		WHERE user_id = $1 AND item_id = $2 AND is_used = false`, userID, itemID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("a tárgy nem található vagy már el lett használva")
	}
	return nil
}

func (r *UserRepository) DeleteInventoryItem(userID uuid.UUID, inventoryID string) error {
	invUUID, err := uuid.Parse(inventoryID)
	if err == nil {
		_, err = r.db.Exec(`DELETE FROM user_inventory WHERE user_id = $1 AND id = $2`, userID, invUUID)
		return err
	}
	_, err = r.db.Exec(`DELETE FROM user_inventory WHERE user_id = $1 AND item_id = $2`, userID, inventoryID)
	return err
}



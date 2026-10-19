package service

import (
	"BootDevClone/services/user-service/internal/domain"
	"github.com/google/uuid"
)

func (s *UserService) GetStoreItems(userID string) ([]domain.StoreItem, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetStoreItems(uid)
}

func (s *UserService) RefreshStore(userID string, cost int) ([]domain.StoreItem, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.RefreshStore(uid, cost)
}

func (s *UserService) PurchaseStoreItem(userID, itemID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.repo.PurchaseItem(uid, itemID)
}

package service

import (
	"context"
	"database/sql"
	"sync"

	"fafhackathon2022/internal/store"
	"fafhackathon2022/internal/store/models"
)

type Service interface {
	GetUser(ctx context.Context, userUUID string) (*models.User, error)
	HandleWSMessage(ctx context.Context, uuid string, msg models.Message) (*models.Message, error)
	//UpdateUserBalance(ctx context.Context, userUUID uint64, amount int) (int, error)
	//GetUsersCoordinates(ctx context.Context, userUUIDs []uint64) ([]redis.GeoLocation, error)
	//GetNearestNCoordinates(ctx context.Context, userUUID string, limit int, location models.Coordinates) ([]models.Coordinates, error)
}

type service struct {
	store store.Store

	roleMutex   *sync.Mutex
	roleSwapper bool
}

func New() Service {
	return &service{
		store:       store.CreateDB(),
		roleMutex:   &sync.Mutex{},
		roleSwapper: false,
	}
}

func (s *service) GetUser(ctx context.Context, userUUID string) (*models.User, error) {
	existingUser, err := s.store.GetUserData(ctx, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			newUser, err := s.store.CreateUser(ctx, userUUID)
			if err != nil {
				return nil, err
			}

			return newUser, nil
		}

	}

	return existingUser, nil
}

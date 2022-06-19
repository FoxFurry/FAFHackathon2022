package store

import (
	"context"
	"fmt"
	"time"

	"fafhackathon2022/internal/store/models"
	"github.com/go-redis/redis/v8"
)

const locationsKey = "locations"

func (s *store) GetUserData(ctx context.Context, userUUID string) (*models.User, error) {
	var u models.User
	if err := s.sql.GetContext(ctx, &u, `SELECT * FROM users WHERE uuid=?`, userUUID); err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *store) CreateUser(ctx context.Context, userUUID string) (*models.User, error) {
	_, err := s.sql.ExecContext(ctx, `INSERT INTO users (uuid) VALUES (?)`, userUUID)
	return &models.User{
		UUID:    userUUID,
		Name:    "User",
		Balance: 500,
	}, err
}

func (s *store) UpdateUserCoordinates(ctx context.Context, location redis.GeoLocation) error {
	if err := s.redis.GeoAdd(ctx, locationsKey, &location).Err(); err != nil {
		return err
	}

	return nil
}

func (s *store) GetNearestNCoordinates(ctx context.Context, limit int, radius float64, location redis.GeoLocation) ([]redis.GeoLocation, error) {
	redisResponse := s.redis.GeoRadius(ctx, locationsKey, location.Longitude, location.Latitude, &redis.GeoRadiusQuery{
		Radius:    radius,
		Unit:      "M",
		WithCoord: true,
		WithDist:  true,
		Count:     limit,
		Sort:      "ASC",
	})

	result, err := redisResponse.Result()
	if err != nil {
		return nil, err
	}

	for idx, element := range result {
		if element.Name == location.Name {
			if idx == len(result)-1 {
				result = result[:idx]
			} else {
				result = append(result[:idx], result[idx+1:]...)
			}
		}
	}

	return result, nil
}

func (s *store) GetUserCoordinates(ctx context.Context, userUUID string) (*redis.GeoLocation, error) {
	result := s.redis.GeoPos(ctx, locationsKey, userUUID)

	if result.Err() != nil {
		return nil, result.Err()
	}

	data, _ := result.Result()

	return &redis.GeoLocation{
		Name:      userUUID,
		Longitude: data[0].Longitude,
		Latitude:  data[0].Latitude,
	}, nil
}

func (s *store) UpdateUserRole(ctx context.Context, userUUID string, userRole models.Role) error {
	if err := s.redis.Set(ctx, userUUID, string(userRole), 10*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

func (s *store) DeleteUserRole(ctx context.Context, userUUID string) error {
	if err := s.redis.Del(ctx, userUUID).Err(); err != nil {
		return err
	}
	return nil
}

func (s *store) GetUserRole(ctx context.Context, userUUID string) (models.Role, error) {
	res := s.redis.Get(ctx, userUUID)
	if res.Err() != nil {
		return "", res.Err()
	}

	role, _ := res.Result()
	switch role {
	case "hunter":
		return models.Hunter, nil
	case "victim":
		return models.Victim, nil
	default:
		return "", fmt.Errorf("unknown role: %s", role)
	}
}

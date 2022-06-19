package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"fafhackathon2022/internal/http/connections"
	"fafhackathon2022/internal/store/models"
	"github.com/go-redis/redis/v8"
)

func (s *service) HandleWSMessage(ctx context.Context, uuid string, msg models.Message) (*models.Message, error) {
	switch msg.Type {
	case models.StartGame:
		return s.HandleStartGame(ctx, uuid, msg)
	case models.Telemetry:
		return s.HandleTelemetry(ctx, uuid, msg)
	case models.Kill:
		return s.HandleKill(ctx, uuid, msg)
	default:
		return nil, fmt.Errorf("unsupported event type: %v", msg.Type)
	}
}

func (s *service) HandleStartGame(ctx context.Context, uuid string, msg models.Message) (*models.Message, error) {
	_, err := s.store.GetUserRole(ctx, uuid)
	if err != redis.Nil {
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("user already has a role")
	}

	var coordinates models.Coordinates
	if err := json.Unmarshal(msg.Data, &coordinates); err != nil {
		return nil, err
	}

	userLocation := redis.GeoLocation{Name: uuid, Latitude: coordinates.Latitude, Longitude: coordinates.Longitude}
	s.store.UpdateUserCoordinates(ctx, userLocation)

	role := s.generateUserModel(ctx, uuid)

	if err := s.store.UpdateUserRole(ctx, uuid, role); err != nil {
		return nil, err
	}

	val, err := json.Marshal(models.StatusUpdate{
		Status:   "start",
		Duration: 600,
		UserRole: role,
	})
	if err != nil {
		return nil, err
	}

	return &models.Message{Type: models.GameStatusUpdate, Data: val}, nil
}

func (s *service) HandleTelemetry(ctx context.Context, uuid string, msg models.Message) (*models.Message, error) {
	var coordinates models.Coordinates
	if err := json.Unmarshal(msg.Data, &coordinates); err != nil {
		return nil, err
	}

	userLocation := redis.GeoLocation{Name: uuid, Latitude: coordinates.Latitude, Longitude: coordinates.Longitude}
	s.store.UpdateUserCoordinates(ctx, userLocation)

	nearby, err := s.store.GetNearestNCoordinates(ctx, 100, 500, userLocation)
	if err != nil {
		return nil, err
	}

	userRole, err := s.store.GetUserRole(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var targetRole models.Role
	if userRole == models.Victim {
		targetRole = models.Hunter
	} else if userRole == models.Hunter {
		targetRole = models.Victim
	} else {
		return nil, fmt.Errorf("unknown user role: %s", userRole)
	}

	var resp []int
	for _, entry := range nearby {
		if nearbyRole, _ := s.store.GetUserRole(ctx, uuid); nearbyRole != targetRole {
			continue
		}

		deltaX := entry.Longitude - coordinates.Longitude
		deltaY := entry.Latitude - coordinates.Latitude
		var radians = math.Atan2(deltaY, deltaX)
		var degree = radians * (180 / math.Pi)
		resp = append(resp, int(degree))
	}

	val, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return &models.Message{Type: models.Enemies, Data: val}, nil
}

func (s *service) HandleKill(ctx context.Context, uuid string, msg models.Message) (*models.Message, error) {
	role, err := s.store.GetUserRole(ctx, uuid)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user doesnt has a role")
		}
		return nil, err
	}
	if role != models.Hunter {
		return nil, fmt.Errorf("user is not a hunter")
	}

	coordinates, err := s.store.GetUserCoordinates(ctx, uuid)
	if err != nil {
		return nil, err
	}

	nearby, err := s.store.GetNearestNCoordinates(ctx, 1000, 50, *coordinates)

	if len(nearby) > 0 {
		err = s.store.DeleteUserRole(ctx, uuid)
		if err != nil {
			return nil, err
		}
	}

	for _, killed := range nearby {
		if client, ok := connections.GetConnection(killed.Name); ok {
			val, err := json.Marshal(models.StatusUpdate{
				Status: "finish",
				Reason: "You were killed",
			})
			if err != nil {
				continue
			}

			err = s.store.DeleteUserRole(ctx, killed.Name)
			if err != nil {
				return nil, err
			}
			client.WriteJSON(&models.Message{Type: models.GameStatusUpdate, Data: val})
		}
	}

	return nil, nil
}

func (s *service) generateUserModel(ctx context.Context, uuid string) models.Role {
	s.roleMutex.Lock()
	defer s.roleMutex.Unlock()

	if s.roleSwapper {
		s.roleSwapper = !s.roleSwapper
		return models.Victim
	} else {
		s.roleSwapper = !s.roleSwapper
		return models.Hunter
	}
}

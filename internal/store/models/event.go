package models

import (
	"encoding/json"
)

type Role string

const (
	Victim Role = "victim"
	Hunter Role = "hunter"
)

type GameStatus string

const (
	Start  GameStatus = "start"
	Finish GameStatus = "finish"
)

type Event string

const (
	StartGame        Event = "startGame"
	GameStatusUpdate Event = "gameStatusUpdate"
	Telemetry        Event = "telemetry"
	Enemies          Event = "enemies"
	Kill             Event = "kill"
)

type Message struct {
	Type Event           `json:"event"`
	Data json.RawMessage `json:"data"`
}

type Coordinates struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type StatusUpdate struct {
	Status   GameStatus `json:"status"`
	Reason   string     `json:"reason,omitempty"`
	Duration int        `json:"duration,omitempty"`
	UserRole Role       `json:"role,omitempty"`
}

type RedisCoordinateHolder struct {
	Name     string
	Location []float64
}

func (r *RedisCoordinateHolder) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r *RedisCoordinateHolder) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

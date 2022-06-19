package models

type User struct {
	ID      uint64 `json:"-"`
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Balance int    `json:"balance"`
}

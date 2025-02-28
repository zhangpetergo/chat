package chatapp

import "github.com/google/uuid"

type user struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

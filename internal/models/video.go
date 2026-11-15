package models

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	Id         uuid.UUID  `json:"id"`
	UserId     uuid.UUID  `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}
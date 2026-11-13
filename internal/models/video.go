package models

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	Id         uuid.UUID
	UserId     uuid.UUID
	Title      string
	Status     string
	CreatedAt  time.Time
}
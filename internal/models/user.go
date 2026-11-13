package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id         uuid.UUID
	Email      string
	CreatedAt  time.Time
}

type OTPReq struct { 
	Email  string  `json:"email"` 
}

type OTPRes struct {
	Email  string  `json:"email"`
	Code   string  `json:"code"`
}
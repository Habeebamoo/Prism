package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	CreatedAt  time.Time  `json:"created_at"`
}

type OTPReq struct { 
	Email  string  `json:"email"` 
}

type OTPRes struct {
	Email  string  `json:"email"`
	Code   string  `json:"code"`
}
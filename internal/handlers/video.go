package handlers

import (
	"database/sql"

	"github.com/Habeebamoo/Prism/internal/configs"
)

type VideoHandler struct {
	cfg  *configs.Config
	DB   *sql.DB
}

func NewVideoHandler(cfg *configs.Config, db *sql.DB) *VideoHandler {
	return &VideoHandler{ cfg: cfg, DB: db }
}

func (v *VideoHandler) DoSomething() {
	
}
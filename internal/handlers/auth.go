package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
	"uuid"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/middlewares"
	"github.com/Habeebamoo/Prism/internal/models"
	"github.com/Habeebamoo/Prism/internal/services"
	"github.com/Habeebamoo/Prism/pkg/utils"
)

type AuthHandler struct {
	cfg   *configs.Config
	DB    *sql.DB
	otps  sync.Map
}

type otpEntry struct {
	Code       string
	ExpiresAt  time.Time
}

func NewAuthHandler(cfg *configs.Config, DB *sql.DB) *AuthHandler {
	return &AuthHandler{cfg: cfg, DB: DB}
}

func (a *AuthHandler) HandleRequestOTP(w http.ResponseWriter, r *http.Request) {
	var req models.OTPReq;

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Invalid JSON Payload" })
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !strings.Contains(email, "@") || len(email) < 5 {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Invalid Email Address" })
		return
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Failed to generate OTP" })
		return
	}
	code := fmt.Sprintf("%06d", n.Int64())

	a.otps.Store(email, otpEntry{
		Code: code,
		ExpiresAt: time.Now().Add(5*time.Minute),
	})

	if err := services.SendOTPEmail(r.Context(), a.cfg, email, code); err != nil {
		log.Printf("[AUTH] email delivery error for %s: %v", email, err)
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Failed to send verification code" })
		return
	}

	utils.JsonResponse(w, 200, utils.Payload{ Success: true, Message: "OTP Sent." })
}

func (a *AuthHandler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req models.OTPRes

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Invalid JSON Payload" })
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	raw, ok := a.otps.Load(email)
	if !ok {
		utils.JsonResponse(w, 401, utils.Payload{ Success: false, Message: "Invalid or Expired OTP" })
		return
	}

	entry := raw.(otpEntry)
	if time.Now().After(entry.ExpiresAt) || entry.Code != code {
		utils.JsonResponse(w, 401, utils.Payload{ Success: false, Message: "Invalid or Expired Code" })
		return
	}
	a.otps.Delete(email)

	var userID string

	err := a.DB.QueryRowContext(
		r.Context(),
		`SELECT id FROM users WHERE email = $1`,
		email,
	).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		userID = uuid.New().String()

		_, err = a.DB.ExecContext(
			r.Context(),
			`INSERT INTO users (id, email) VALUES ($1, $2)`,
			userID,
			email,
		)

		if err != nil {
			utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Account Creation Failed" })
			return
		}
	} else if err != nil {
		log.Println(err)
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Internal Server Error" })
		return
	}

	token, err := utils.SignJWT(a.cfg, userID)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	var sameSite http.SameSite

	if a.cfg.Env == "production" {
		sameSite = http.SameSiteNoneMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "prism_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   a.cfg.Env == "production",
		MaxAge:   7 * 24 * 60 * 60 * 1000,
	})

	utils.JsonResponse(w, 200, utils.Payload{ Success: true, Message: "Authentication Successfull" })
}

func (a *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "prism_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	utils.JsonResponse(w, 200, utils.Payload{ Success: false, Message: "Logged out successfully" })
}

func (a *AuthHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		utils.JsonResponse(w, http.StatusUnauthorized, utils.Payload{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var user models.User
	err := a.DB.QueryRowContext(
		r.Context(),
		`SELECT id, email, created_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.Id, &user.Email, &user.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		utils.JsonResponse(w, http.StatusNotFound, utils.Payload{
			Success: false,
			Message: "User not found",
		})
		return
	}
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, utils.Payload{
			Success: false,
			Message: "Failed to fetch user",
		})
		return
	}

	utils.JsonResponse(w, http.StatusOK, utils.Payload{
		Success: true,
		Data:    user,
	})
}
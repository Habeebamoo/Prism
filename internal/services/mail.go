package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
)

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

func SendOTPEmail(ctx context.Context, cfg *configs.Config, toEmail, otp string) error {
	if cfg.ResendAPIKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not configured")
	}

	from := "prism@myclivo.com"

	payload := resendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Your Prism login code: %s", otp),
		Text:    fmt.Sprintf("Your verification code is %s.\n\nIt expires in 5 minutes. If you didn't request this, ignore this email.", otp),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error (%d): %s", resp.StatusCode, string(errBody))
	}

	return nil
}
package middlewares

import (
	"context"
	"net/http"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/pkg/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(cfg *configs.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("prism_token")
			if err != nil || cookie.Value == "" {
				utils.JsonResponse(w, http.StatusUnauthorized, utils.Payload{
					Success: false,
					Message: "Unauthorized: missing authentication token",
				})
				return
			}

			claims, err := utils.ParseJWT(cfg, cookie.Value)
			if err != nil {
				utils.JsonResponse(w, http.StatusUnauthorized, utils.Payload{
					Success: false,
					Message: "Unauthorized: invalid or expired token",
				})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(cfg *configs.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("prism_token")

			if err == nil && cookie.Value != "" {
				if claims, err := utils.ParseJWT(cfg, cookie.Value); err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
					r = r.WithContext(ctx)
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok && userID != ""
}
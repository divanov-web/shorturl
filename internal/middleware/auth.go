package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	authCookieName = "auth_token"
)

type contextKey string

const userIDKey contextKey = "user_id"

type Auth struct {
	Secret string
}

func NewAuth(secret string) *Auth {
	return &Auth{Secret: secret}
}

func (a *Auth) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(authCookieName)
		if err == nil {
			token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.Secret), nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if val, ok := claims["user_id"].(string); ok {
						userID = val
					}
				}
			}
		}

		if userID == "" {
			userID = uuid.NewString()
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user_id": userID,
				"exp":     time.Now().Add(365 * 24 * time.Hour).Unix(),
			})
			signed, err := token.SignedString([]byte(a.Secret))
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     authCookieName,
					Value:    signed,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает user_id из context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

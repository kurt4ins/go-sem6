package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kurt4ins/taskmanager/internal/utils"
)

type contextKey string

const userIdKey contextKey = "userId"

func UserIdFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIdKey).(int)
	return id, ok
}

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return secret, nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					utils.WriteError(w, http.StatusUnauthorized, "token expired")
				} else {
					utils.WriteError(w, http.StatusUnauthorized, "invalid token")
				}
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			sub, ok := claims["sub"].(float64)
			if !ok {
				utils.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			if claims["type"] != "access" {
				utils.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIdKey, int(sub))
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

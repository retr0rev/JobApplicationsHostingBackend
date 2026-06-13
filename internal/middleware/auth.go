package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ClientIDKey contextKey = "client_id"
	AdminIDKey  contextKey = "admin_id"
)

func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func GenerateToken(claims jwt.MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(72 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func requireAuth(requiredRole string, idClaim string, key contextKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return getJWTSecret(), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			role, _ := claims["role"].(string)
			if role != requiredRole {
				http.Error(w, `{"error":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			idFloat, ok := claims[idClaim].(float64)
			if !ok {
				http.Error(w, `{"error":"invalid token payload"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), key, int64(idFloat))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClientAuth(next http.Handler) http.Handler {
	return requireAuth("client", "client_id", ClientIDKey)(next)
}

func AdminAuth(next http.Handler) http.Handler {
	return requireAuth("admin", "admin_id", AdminIDKey)(next)
}

func GetClientID(r *http.Request) int64 {
	id, _ := r.Context().Value(ClientIDKey).(int64)
	return id
}

func GetAdminID(r *http.Request) int64 {
	id, _ := r.Context().Value(AdminIDKey).(int64)
	return id
}

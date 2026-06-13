package middleware

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidatePassword(password string) string {
	if len(password) < 8 {
		return "password must be at least 8 characters"
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return "password must contain uppercase, lowercase, and digit characters"
	}
	return ""
}

func ValidateJobTitle(title string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "job_title is required"
	}
	if len(trimmed) > 200 {
		return "job_title must be under 200 characters"
	}
	return ""
}

func ValidateDescription(desc string) string {
	trimmed := strings.TrimSpace(desc)
	if trimmed == "" {
		return "description is required"
	}
	if len(trimmed) > 5000 {
		return "description must be under 5000 characters"
	}
	return ""
}

func ValidateCategory(cat string) string {
	if len(cat) > 100 {
		return "category must be under 100 characters"
	}
	return ""
}

func ValidateLocation(loc string) string {
	if len(loc) > 200 {
		return "location must be under 200 characters"
	}
	return ""
}

func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin != "" && origin != "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		if os.Getenv("TLS_ENABLED") == "true" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func CheckJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("CRITICAL: JWT_SECRET environment variable must be set to a unique secret key")
	}
	if len(secret) < 32 {
		log.Fatal("CRITICAL: JWT_SECRET must be at least 32 characters (recommend 64 hex chars from `openssl rand -hex 32`)")
	}
}

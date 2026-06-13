package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"jobapps/internal/email"
	"jobapps/internal/handlers"
	authmw "jobapps/internal/middleware"
	"jobapps/internal/repository"
	dbpkg "jobapps/pkg/database"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	authmw.CheckJWTSecret()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./database/database.db"
	}

	db, err := dbpkg.NewDB(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if len(os.Args) >= 2 && os.Args[1] == "seed-admin" {
		runSeed(db)
		return
	}

	var emailSender email.Sender
	if os.Getenv("SMTP_HOST") != "" {
		emailSender = email.NewSMTPSender()
		log.Println("email: using SMTP")
	} else {
		emailSender = email.NewConsoleSender()
		log.Println("email: using console (dev mode)")
	}

	clientRepo := repository.NewClientRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	jobRepo := repository.NewJobRepo(db)

	companyHandler := handlers.NewCompanyHandler(clientRepo, emailSender)
	verifyHandler := handlers.NewVerifyHandler(clientRepo, emailSender)
	adminHandler := handlers.NewAdminHandler(adminRepo, jobRepo)
	jobHandler := handlers.NewJobHandler(jobRepo)

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		log.Println("WARNING: CORS_ORIGIN not set, defaulting to '*' (insecure for production). Set CORS_ORIGIN to a specific origin.")
		corsOrigin = "*"
	} else if corsOrigin == "*" {
		log.Println("WARNING: CORS_ORIGIN is set to '*'. This disables credentials and is not recommended for production.")
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(authmw.CORS(corsOrigin))
	r.Use(authmw.SecurityHeaders)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/api/auth/signup", companyHandler.Signup)
	r.Post("/api/auth/login", companyHandler.Login)
	r.Get("/api/auth/verify", verifyHandler.VerifyEmail)
	r.Post("/api/auth/forgot-password", verifyHandler.ForgotPassword)
	r.Post("/api/auth/reset-password", verifyHandler.ResetPassword)

	r.Route("/api/jobs", func(r chi.Router) {
		r.Use(authmw.ClientAuth)
		r.Get("/", jobHandler.List)
		r.Post("/", jobHandler.Create)
		r.Get("/{id}", jobHandler.Get)
		r.Put("/{id}", jobHandler.Update)
		r.Delete("/{id}", jobHandler.Delete)
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", adminHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authmw.AdminAuth)
			r.Get("/jobs", adminHandler.ListJobs)
			r.Put("/jobs/{id}/status", adminHandler.UpdateStatus)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}

func runSeed(db *sql.DB) {
	cmd := flag.NewFlagSet("seed-admin", flag.ExitOnError)
	emailFlag := cmd.String("email", "", "admin email")
	passwordFlag := cmd.String("password", "", "admin password")
	cmd.Parse(os.Args[2:])

	if *emailFlag == "" || *passwordFlag == "" {
		log.Fatal("usage: go run ./cmd/server seed-admin --email=admin@site.com --password=securepass")
	}

	email := strings.ToLower(strings.TrimSpace(*emailFlag))

	hash, err := bcrypt.GenerateFromPassword([]byte(*passwordFlag), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}

	_, err = db.Exec("INSERT INTO ADMINS (email, password) VALUES (?, ?)", email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Fatalf("admin already exists: %s (use a different email or remove the existing admin)", email)
		}
		log.Fatalf("seed: %v", err)
	}
	log.Printf("admin seeded: %s", email)
}

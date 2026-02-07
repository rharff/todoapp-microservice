package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"audit-service/internal/db"
	"audit-service/internal/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := connectWithRetry(dbURL, 30*time.Second)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.EnsureSchema(ctx, pool); err != nil {
		log.Fatalf("schema setup failed: %v", err)
	}

	h := httpapi.NewHandler(pool)

	r := chi.NewRouter()
	r.Use(
		httpapi.CORS,
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)
	r.Mount("/", h.Routes())

	server := &http.Server{
		Addr:              ":8081",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Audit Service running on :8081")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func connectWithRetry(dbURL string, maxWait time.Duration) (*pgxpool.Pool, error) {
	deadline := time.Now().Add(maxWait)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err := pgxpool.New(ctx, dbURL)
		cancel()
		if err == nil {
			return pool, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(2 * time.Second)
	}
}

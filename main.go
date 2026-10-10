package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type pinger interface {
	Ping(ctx context.Context) error
}

func healthz(logger *slog.Logger, db pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		code := http.StatusOK
		body := map[string]string{"status": "ok", "db": "ok"}
		if err := db.Ping(ctx); err != nil {
			logger.Error("pengecekan database gagal", "error", err)
			code = http.StatusServiceUnavailable
			body = map[string]string{"status": "degraded", "db": "error"}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(body)
	}
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

func newRouter(logger *slog.Logger, db pinger) http.Handler {
	r := chi.NewRouter()
	r.Use(requestLogger(logger))
	r.Get("/healthz", healthz(logger, db))
	return r
}

func connectDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL tidak valid: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("tidak bisa terhubung ke database (sudah menyala? coba: soalku-db start): %w", err)
	}
	return pool, nil
}

func run(logger *slog.Logger) error {
	// Membaca .env bila ada; variabel yang sudah diset di terminal tidak ditimpa.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("DATABASE_URL belum diatur (salin .env.example menjadi .env)")
	}
	pool, err := connectDB(context.Background(), dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("terhubung ke database")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	logger.Info("server berjalan", "addr", addr)

	return http.ListenAndServe(addr, newRouter(logger, pool))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server berhenti", "error", err)
		os.Exit(1)
	}
}

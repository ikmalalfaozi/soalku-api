package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeDB struct{ err error }

func (f fakeDB) Ping(ctx context.Context) error { return f.err }

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func getHealthz(t *testing.T, db pinger) (*httptest.ResponseRecorder, map[string]string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthz(discardLogger(), db).ServeHTTP(rec, req)

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("isi respons bukan JSON valid: %v", err)
	}
	return rec, body
}

func TestHealthzDatabaseSehat(t *testing.T) {
	rec, body := getHealthz(t, fakeDB{})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, harusnya %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, harusnya application/json", got)
	}
	if body["status"] != "ok" || body["db"] != "ok" {
		t.Errorf(`body = %v, harusnya status "ok" dan db "ok"`, body)
	}
}

func TestHealthzDatabaseMati(t *testing.T) {
	rec, body := getHealthz(t, fakeDB{err: errors.New("koneksi putus")})

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, harusnya %d", rec.Code, http.StatusServiceUnavailable)
	}
	if body["status"] != "degraded" || body["db"] != "error" {
		t.Errorf(`body = %v, harusnya status "degraded" dan db "error"`, body)
	}
}

func TestRouter(t *testing.T) {
	tests := []struct {
		nama   string
		method string
		path   string
		want   int
	}{
		{"GET /healthz diterima", http.MethodGet, "/healthz", http.StatusOK},
		{"POST /healthz ditolak", http.MethodPost, "/healthz", http.StatusMethodNotAllowed},
		{"alamat tidak dikenal", http.MethodGet, "/tidak-ada", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			newRouter(discardLogger(), fakeDB{}).ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, harusnya %d", rec.Code, tc.want)
			}
		})
	}
}

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	newRouter(logger, fakeDB{}).ServeHTTP(rec, req)

	log := buf.String()
	for _, want := range []string{"method=GET", "path=/healthz", "status=200", "duration_ms="} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log tidak memuat %q, isi log: %s", want, log)
		}
	}
}

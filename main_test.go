package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, harusnya %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, harusnya application/json", got)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("isi respons bukan JSON valid: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf(`status = %q, harusnya "ok"`, body["status"])
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

			newRouter(slog.New(slog.NewTextHandler(io.Discard, nil))).ServeHTTP(rec, req)

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
	newRouter(logger).ServeHTTP(rec, req)

	log := buf.String()
	for _, want := range []string{"method=GET", "path=/healthz", "status=200", "duration_ms="} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("log tidak memuat %q, isi log: %s", want, log)
		}
	}
}

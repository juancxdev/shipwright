package harness

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateStitchConnectionPostsMCPInitialize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.Header.Get("X-Goog-Api-Key") != "abc123" {
			t.Fatalf("api key header = %q", r.Header.Get("X-Goog-Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-03-26"}}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := ValidateStitchConnection(ctx, server.Client(), server.URL, "abc123")
	if !result.Checked || !result.Healthy || result.Status != "healthy" {
		t.Fatalf("ValidateStitchConnection = %+v", result)
	}
}

func TestValidateStitchConnectionReportsAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := ValidateStitchConnection(ctx, server.Client(), server.URL, "bad")
	if !result.Checked || result.Healthy || result.Status != "auth_failed" {
		t.Fatalf("ValidateStitchConnection = %+v", result)
	}
}

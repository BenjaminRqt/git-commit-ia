package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"git-ai-commit/internal/ai/openai"
)

func newTestAdapter(t *testing.T, baseURL string) *openai.Adapter {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", "test-key")
	a, err := openai.New(baseURL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func TestComplete_ParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "feat - Add login - VAL-1"}},
			},
		})
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	got, err := a.Complete(context.Background(), "system", "user", openai.DefaultModel, 1024)
	if err != nil {
		t.Fatalf("Complete() unexpected error: %v", err)
	}
	if got != "feat - Add login - VAL-1" {
		t.Errorf("Complete() = %q", got)
	}
}

func TestComplete_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid api key"},
		})
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Complete(context.Background(), "system", "user", openai.DefaultModel, 1024)
	if err == nil {
		t.Error("Complete() should return an error for a 401 status")
	}
}

func TestNew_MissingKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	_, err := openai.New("")
	if err == nil {
		t.Error("New() should return an error when OPENAI_API_KEY is missing")
	}
}

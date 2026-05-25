package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	lastCtx context.Context
	lastN   int

	queries []string
	err     error
}

func (m *mockService) TopNQueries(ctx context.Context, n int) ([]string, error) {
	m.lastCtx = ctx
	m.lastN = n
	return m.queries, m.err
}

func TestGetTopNQueriesMissingN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockService{}
	handlers := NewHandlers(svc)

	router := gin.New()
	router.GET("/top-requests", handlers.GetTopNQueries)

	req := httptest.NewRequest(http.MethodGet, "/top-requests", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetTopNQueriesInvalidN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockService{}
	handlers := NewHandlers(svc)

	router := gin.New()
	router.GET("/top-requests", handlers.GetTopNQueries)

	req := httptest.NewRequest(http.MethodGet, "/top-requests?n=0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetTopNQueriesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockService{queries: []string{"alpha", "beta"}}
	handlers := NewHandlers(svc)

	router := gin.New()
	router.GET("/top-requests", handlers.GetTopNQueries)

	req := httptest.NewRequest(http.MethodGet, "/top-requests?n=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload TopNResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(payload.Requests) != 2 || payload.Requests[0] != "alpha" || payload.Requests[1] != "beta" {
		t.Fatalf("unexpected response: %+v", payload.Requests)
	}

	if svc.lastN != 2 {
		t.Fatalf("expected n=2, got %d", svc.lastN)
	}
	if svc.lastCtx == nil {
		t.Fatalf("expected context to be passed to service")
	}
}

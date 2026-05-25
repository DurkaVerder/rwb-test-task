package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	customErr "github.com/DurkaVerder/rwb-test-task/internal/errors"
	"github.com/gin-gonic/gin"
)

type mockStopListService struct {
	lastCtx  context.Context
	lastWord string

	stopWords []string
	errGet    error
	errAdd    error
	errRemove error
}

func (m *mockStopListService) GetAllStopWords(ctx context.Context) ([]string, error) {
	m.lastCtx = ctx
	return m.stopWords, m.errGet
}

func (m *mockStopListService) AddStopWord(ctx context.Context, word string) error {
	m.lastCtx = ctx
	m.lastWord = word
	return m.errAdd
}

func (m *mockStopListService) RemoveStopWord(ctx context.Context, word string) error {
	m.lastCtx = ctx
	m.lastWord = word
	return m.errRemove
}

func TestGetStopListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{stopWords: []string{"spam"}}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.GET("/stoplist", handlers.GetStopList)

	req := httptest.NewRequest(http.MethodGet, "/stoplist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload StopListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(payload.StopWords) != 1 || payload.StopWords[0] != "spam" {
		t.Fatalf("unexpected response: %+v", payload.StopWords)
	}
	if svc.lastCtx == nil {
		t.Fatalf("expected context to be passed to service")
	}
}

func TestAddStopWordInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.POST("/stoplist", handlers.AddStopWord)

	req := httptest.NewRequest(http.MethodPost, "/stoplist", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAddStopWordInvalidWord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{errAdd: customErr.InvalidWordError}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.POST("/stoplist", handlers.AddStopWord)

	req := httptest.NewRequest(http.MethodPost, "/stoplist", bytes.NewBufferString(`{"word":" "}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAddStopWordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.POST("/stoplist", handlers.AddStopWord)

	req := httptest.NewRequest(http.MethodPost, "/stoplist", bytes.NewBufferString(`{"word":"spam"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRemoveStopWordNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{errRemove: customErr.WordNotFoundError}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.DELETE("/stoplist", handlers.RemoveStopWord)

	req := httptest.NewRequest(http.MethodDelete, "/stoplist", bytes.NewBufferString(`{"word":"spam"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRemoveStopWordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockStopListService{}
	handlers := NewStopListHandlers(svc)

	router := gin.New()
	router.DELETE("/stoplist", handlers.RemoveStopWord)

	req := httptest.NewRequest(http.MethodDelete, "/stoplist", bytes.NewBufferString(`{"word":"spam"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

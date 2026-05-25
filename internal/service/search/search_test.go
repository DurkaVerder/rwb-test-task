package search

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepo struct {
	lastCtx   context.Context
	lastN     int
	lastQuery string
	lastAt    time.Time

	queries []string
	err     error
}

func (m *mockRepo) GetTopNQueries(ctx context.Context, n int) ([]string, error) {
	m.lastCtx = ctx
	m.lastN = n
	return m.queries, m.err
}

func (m *mockRepo) AddQuery(ctx context.Context, query string, at time.Time) error {
	m.lastCtx = ctx
	m.lastQuery = query
	m.lastAt = at
	return m.err
}

func TestSearchServiceTopNQueries(t *testing.T) {
	repo := &mockRepo{queries: []string{"alpha", "beta"}}
	svc := NewSearchService(repo)

	ctx := context.Background()
	result, err := svc.TopNQueries(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCtx != ctx {
		t.Fatalf("expected ctx to be passed to repo")
	}
	if repo.lastN != 2 {
		t.Fatalf("expected n=2, got %d", repo.lastN)
	}
	if len(result) != 2 || result[0] != "alpha" || result[1] != "beta" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSearchServiceTopNQueriesError(t *testing.T) {
	repo := &mockRepo{err: errors.New("boom")}
	svc := NewSearchService(repo)

	_, err := svc.TopNQueries(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSearchServiceAddQuery(t *testing.T) {
	repo := &mockRepo{}
	svc := NewSearchService(repo)

	at := time.Unix(123, 0)
	ctx := context.Background()
	if err := svc.AddQuery(ctx, "gamma", at); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastCtx != ctx {
		t.Fatalf("expected ctx to be passed to repo")
	}
	if repo.lastQuery != "gamma" {
		t.Fatalf("expected query to be captured")
	}
	if !repo.lastAt.Equal(at) {
		t.Fatalf("expected at to be captured")
	}
}

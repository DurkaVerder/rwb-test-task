package stoplist

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	lastCtx  context.Context
	lastWord string

	stopWords []string
	errAdd    error
	errRemove error
	errGet    error
}

func (m *mockRepo) AddStopWord(ctx context.Context, word string) error {
	m.lastCtx = ctx
	m.lastWord = word
	return m.errAdd
}

func (m *mockRepo) RemoveStopWord(ctx context.Context, word string) error {
	m.lastCtx = ctx
	m.lastWord = word
	return m.errRemove
}

func (m *mockRepo) GetAllStopWords(ctx context.Context) ([]string, error) {
	m.lastCtx = ctx
	return m.stopWords, m.errGet
}

func TestStopListServiceAddStopWord(t *testing.T) {
	repo := &mockRepo{}
	svc := NewStopListService(repo)

	ctx := context.Background()
	if err := svc.AddStopWord(ctx, "spam"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCtx != ctx {
		t.Fatalf("expected ctx to be passed to repo")
	}
	if repo.lastWord != "spam" {
		t.Fatalf("expected word to be passed to repo")
	}
}

func TestStopListServiceAddStopWordError(t *testing.T) {
	repo := &mockRepo{errAdd: errors.New("boom")}
	svc := NewStopListService(repo)

	if err := svc.AddStopWord(context.Background(), "spam"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestStopListServiceRemoveStopWord(t *testing.T) {
	repo := &mockRepo{}
	svc := NewStopListService(repo)

	ctx := context.Background()
	if err := svc.RemoveStopWord(ctx, "spam"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCtx != ctx {
		t.Fatalf("expected ctx to be passed to repo")
	}
	if repo.lastWord != "spam" {
		t.Fatalf("expected word to be passed to repo")
	}
}

func TestStopListServiceRemoveStopWordError(t *testing.T) {
	repo := &mockRepo{errRemove: errors.New("boom")}
	svc := NewStopListService(repo)

	if err := svc.RemoveStopWord(context.Background(), "spam"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestStopListServiceGetAllStopWords(t *testing.T) {
	repo := &mockRepo{stopWords: []string{"a", "b"}}
	svc := NewStopListService(repo)

	ctx := context.Background()
	words, err := svc.GetAllStopWords(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCtx != ctx {
		t.Fatalf("expected ctx to be passed to repo")
	}
	if len(words) != 2 || words[0] != "a" || words[1] != "b" {
		t.Fatalf("unexpected stop words: %+v", words)
	}
}

func TestStopListServiceGetAllStopWordsError(t *testing.T) {
	repo := &mockRepo{errGet: errors.New("boom")}
	svc := NewStopListService(repo)

	if _, err := svc.GetAllStopWords(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

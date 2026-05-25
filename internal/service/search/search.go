package search

import (
	"context"
	"time"
)

type Repository interface {
	GetTopNQueries(ctx context.Context, n int) ([]string, error)
	AddQuery(ctx context.Context, query string, at time.Time) error
}

type SearchService struct {
	repo Repository
}

func NewSearchService(repo Repository) *SearchService {
	return &SearchService{
		repo: repo,
	}
}

func (s *SearchService) TopNQueries(ctx context.Context, n int) ([]string, error) {
	queries, err := s.repo.GetTopNQueries(ctx, n)
	if err != nil {
		return nil, err
	}
	return queries, nil
}

func (s *SearchService) AddQuery(ctx context.Context, query string, at time.Time) error {
	if err := s.repo.AddQuery(ctx, query, at); err != nil {
		return err
	}
	return nil
}

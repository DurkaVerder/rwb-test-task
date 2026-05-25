package service

import (
	"context"
	"time"
)

type Repository interface {
	GetTopNQueries(ctx context.Context, n int) ([]string, error)
	AddQuery(ctx context.Context, query string, at time.Time) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) TopNQueries(ctx context.Context, n int) ([]string, error) {
	queries, err := s.repo.GetTopNQueries(ctx, n)
	if err != nil {
		return nil, err
	}
	return queries, nil
}

func (s *Service) AddQuery(ctx context.Context, query string, at time.Time) error {
	if err := s.repo.AddQuery(ctx, query, at); err != nil {
		return err
	}
	return nil
}

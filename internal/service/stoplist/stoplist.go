package stoplist

import "context"

type StopWordsRepository interface {
	AddStopWord(ctx context.Context, word string) error
	RemoveStopWord(ctx context.Context, word string) error
	GetAllStopWords(ctx context.Context) ([]string, error)
}

type StopListService struct {
	repo StopWordsRepository
}

func NewStopListService(repo StopWordsRepository) *StopListService {
	return &StopListService{
		repo: repo,
	}
}

func (s *StopListService) AddStopWord(ctx context.Context, word string) error {
	if err := s.repo.AddStopWord(ctx, word); err != nil {
		return err
	}
	return nil
}

func (s *StopListService) RemoveStopWord(ctx context.Context, word string) error {
	if err := s.repo.RemoveStopWord(ctx, word); err != nil {
		return err
	}
	return nil
}

func (s *StopListService) GetAllStopWords(ctx context.Context) ([]string, error) {
	stopWords, err := s.repo.GetAllStopWords(ctx)
	if err != nil {
		return nil, err
	}
	return stopWords, nil
}

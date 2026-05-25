package service

type Repository interface {
	GetTopNRequests(n int) ([]string, error)
	AddRequest(request string) error
}

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		Repo: repo,
	}
}

func (s *Service) TopNRequests(n int) ([]string, error) {
	requests, err := s.Repo.GetTopNRequests(n)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *Service) AddRequest(request string) error {
	if err := s.Repo.AddRequest(request); err != nil {
		return err
	}
	return nil
}

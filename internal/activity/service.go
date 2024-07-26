package activity

import "context"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s Service) CreateActivity(ctx context.Context, req Activity) error {
	return s.repo.createActivity(ctx, req)
}

func (s Service) ListActivity(ctx context.Context, req FilterActivity) (res ActivityList, err error) {
	return s.repo.listActivities(ctx, req)
}

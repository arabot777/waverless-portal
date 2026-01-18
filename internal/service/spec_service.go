package service

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
)

type SpecService struct {
	repo *mysql.SpecRepo
}

func NewSpecService(repo *mysql.SpecRepo) *SpecService {
	return &SpecService{repo: repo}
}

func (s *SpecService) ListSpecs(ctx context.Context, specType string) ([]mysql.SpecWithAvailability, error) {
	return s.repo.ListWithAvailability(ctx, specType)
}

func (s *SpecService) GetSpec(ctx context.Context, specName string) (*model.SpecPricing, error) {
	return s.repo.GetByName(ctx, specName)
}

func (s *SpecService) ListAll(ctx context.Context) ([]model.SpecPricing, error) {
	return s.repo.List(ctx)
}

func (s *SpecService) CreateSpec(ctx context.Context, spec *model.SpecPricing) error {
	return s.repo.Create(ctx, spec)
}

func (s *SpecService) UpdateSpec(ctx context.Context, id int64, updates map[string]interface{}) error {
	return s.repo.Update(ctx, id, updates)
}

func (s *SpecService) DeleteSpec(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *SpecService) EstimateCost(ctx context.Context, specName string, hours float64, replicas int) (int64, error) {
	spec, err := s.repo.GetByName(ctx, specName)
	if err != nil {
		return 0, err
	}
	return int64(float64(spec.PricePerHour) * hours * float64(replicas)), nil
}

package service

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
)

type BillingService struct {
	repo         *mysql.BillingRepo
	userRepo     *mysql.UserRepo
	endpointRepo *mysql.EndpointRepo
}

func NewBillingService(repo *mysql.BillingRepo, userRepo *mysql.UserRepo, endpointRepo *mysql.EndpointRepo) *BillingService {
	return &BillingService{repo: repo, userRepo: userRepo, endpointRepo: endpointRepo}
}

func (s *BillingService) GetUsageStats(ctx context.Context, userID string, from, to time.Time) (map[string]interface{}, error) {
	totalAmount, totalSeconds, err := s.repo.GetUsageStats(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_amount":  totalAmount,
		"total_seconds": totalSeconds,
		"from":          from,
		"to":            to,
	}, nil
}

func (s *BillingService) GetWorkerBillingRecords(ctx context.Context, userID string, limit, offset int) ([]mysql.BillingTransactionWithEndpoint, int64, error) {
	return s.repo.ListTransactions(ctx, userID, limit, offset)
}

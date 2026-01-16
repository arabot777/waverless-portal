package service

import (
	"context"
	"errors"

	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type UserService struct {
	repo *mysql.UserRepo
}

func NewUserService(repo *mysql.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) EnsureUserBalance(ctx context.Context, userID, orgID, userName, email string) (*model.UserBalance, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err == nil {
		if balance.OrgID != orgID || balance.UserName != userName || balance.Email != email {
			s.repo.UpdateBalance(ctx, userID, map[string]interface{}{
				"org_id": orgID, "user_name": userName, "email": email,
			})
		}
		return balance, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	balance = &model.UserBalance{
		UserID: userID, OrgID: orgID, UserName: userName, Email: email,
		Balance: 0, Currency: "USD", Status: "active",
	}
	if err := s.repo.CreateBalance(ctx, balance); err != nil {
		return nil, err
	}
	logger.InfoCtx(ctx, "Created user balance record for user: %s", userID)
	return balance, nil
}

func (s *UserService) GetBalance(ctx context.Context, userID string) (*model.UserBalance, error) {
	return s.repo.GetBalance(ctx, userID)
}

func (s *UserService) DeductBalance(ctx context.Context, userID string, amount float64) error {
	return s.repo.Transaction(ctx, func(tx *mysql.UserRepo) error {
		balance, err := tx.GetBalance(ctx, userID)
		if err != nil {
			return err
		}
		if balance.Balance < amount {
			return errors.New("insufficient balance")
		}
		return tx.DeductBalance(ctx, userID, amount)
	})
}

func (s *UserService) AddBalance(ctx context.Context, userID string, amount float64) error {
	return s.repo.AddBalance(ctx, userID, amount)
}

func (s *UserService) CreateRechargeRecord(ctx context.Context, record *model.RechargeRecord) error {
	return s.repo.CreateRechargeRecord(ctx, record)
}

func (s *UserService) CompleteRecharge(ctx context.Context, recordID int64) error {
	return s.repo.Transaction(ctx, func(tx *mysql.UserRepo) error {
		record, err := tx.GetRechargeRecord(ctx, recordID)
		if err != nil {
			return err
		}
		if err := tx.UpdateRechargeRecord(ctx, recordID, map[string]interface{}{
			"status": "completed", "completed_at": gorm.Expr("NOW()"),
		}); err != nil {
			return err
		}
		return tx.AddBalance(ctx, record.UserID, record.Amount)
	})
}

func (s *UserService) GetRechargeRecords(ctx context.Context, userID string, limit, offset int) ([]model.RechargeRecord, int64, error) {
	return s.repo.ListRechargeRecords(ctx, userID, limit, offset)
}

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

func (s *UserService) EnsureUser(ctx context.Context, userID, orgID, userName, email string) (*model.UserBalance, error) {
	user, err := s.repo.GetBalance(ctx, userID)
	if err == nil {
		if user.OrgID != orgID || user.UserName != userName || user.Email != email {
			s.repo.UpdateBalance(ctx, userID, map[string]interface{}{
				"org_id": orgID, "user_name": userName, "email": email,
			})
		}
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user = &model.UserBalance{
		UserID: userID, OrgID: orgID, UserName: userName, Email: email,
	}
	if err := s.repo.CreateBalance(ctx, user); err != nil {
		return nil, err
	}
	logger.InfoCtx(ctx, "Created user record for user: %s", userID)
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*model.UserBalance, error) {
	return s.repo.GetBalance(ctx, userID)
}

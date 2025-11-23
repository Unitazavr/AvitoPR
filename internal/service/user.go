package service

import (
	"context"
	"errors"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/jackc/pgx/v5"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetUserReviews(ctx context.Context, userID string) (*domain.UserReport, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "user not found",
				},
			}
		}
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserReviews(ctx context.Context, userID string) (*domain.UserReport, error) {
	prs, err := s.userRepo.GetPullRequests(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в UserReport
	pullRequests := make([]domain.PullRequest, 0, len(prs))
	for _, pr := range prs {
		pullRequests = append(pullRequests, domain.PullRequest{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		})
	}

	return &domain.UserReport{
		UserID:       userID,
		PullRequests: pullRequests,
	}, nil
}

package service

import (
	"context"
	"errors"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

type PrService interface {
	CreatePR(ctx context.Context, pr *domain.PullRequestShort) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignPR(ctx context.Context, pullRequestID, oldUserID string) (pr *domain.PullRequest, newReviewerID string, err error)
}

type prService struct {
	prRepo repository.PrRepository
}

func NewPrService(prRepo repository.PrRepository) PrService {
	return &prService{
		prRepo: prRepo,
	}
}

func (s *prService) CreatePR(ctx context.Context, pr *domain.PullRequestShort) (*domain.PullRequest, error) {
	err := s.prRepo.Create(ctx, pr)
	if err != nil {
		// Проверяем ошибку уникальности (PR уже существует)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodePRExists,
					Message: "PR id already exists",
				},
			}
		}

		// Проверяем ошибку NOT NULL или FOREIGN KEY (автор не найден)
		if errors.As(err, &pgErr) && (pgErr.Code == "23502" || pgErr.Code == "23503") {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "author or team not found",
				},
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "author or team not found",
				},
			}
		}

		return nil, err
	}

	// Возвращаем полный объект PullRequest
	return &domain.PullRequest{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: []string{}, // Будет заполнено позже, если нужно
	}, nil
}

func (s *prService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	err := s.prRepo.Merge(ctx, prID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already merged") {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "PR not found",
				},
			}
		}
		return nil, err
	}

	// Возвращаем обновленный PR
	return &domain.PullRequest{
		PullRequestID: prID,
		Status:        domain.PRStatusMerged,
	}, nil
}

func (s *prService) ReassignPR(ctx context.Context, pullRequestID, oldUserID string) (*domain.PullRequest, string, error) {
	err := s.prRepo.Reassign(ctx, pullRequestID, oldUserID)
	if err != nil {
		// PR не найден
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "PR or user not found",
				},
			}
		}

		// Проверяем специфичные ошибки по тексту сообщения
		errMsg := err.Error()

		if strings.Contains(errMsg, "cannot reassign reviewers for merged PR") {
			return nil, "", &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodePRMerged,
					Message: "cannot reassign on merged PR",
				},
			}
		}

		if strings.Contains(errMsg, "user is not a reviewer of this PR") {
			return nil, "", &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			}
		}

		if strings.Contains(errMsg, "no available reviewers in team") {
			return nil, "", &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNoCandidate,
					Message: "no active replacement candidate in team",
				},
			}
		}

		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "not in any team") {
			return nil, "", &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "PR or user not found",
				},
			}
		}

		return nil, "", err
	}

	// Возвращаем обновленный PR (в реальности нужно было бы получить новый список ревьюверов)
	return &domain.PullRequest{
		PullRequestID: pullRequestID,
		Status:        domain.PRStatusOpen,
	}, "", nil // новый reviewer_id нужно получить из репозитория
}

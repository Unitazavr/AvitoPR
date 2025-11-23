package service

import (
	"context"
	"errors"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/Unitazavr/AvitoPR/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeamByName(ctx context.Context, name string) (*domain.Team, error)
}

type teamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	err := s.teamRepo.Create(ctx, team)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeTeamExists,
					Message: "team already exists",
				},
			}
		}
		return nil, err
	}
	team, err = s.GetTeamByName(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) GetTeamByName(ctx context.Context, name string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.ErrorResponse{
				ErrorContent: domain.ErrorBody{
					Code:    domain.ErrCodeNotFound,
					Message: "team not found",
				},
			}
		}
		return nil, err
	}

	return team, nil
}

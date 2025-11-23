package repository

import (
	"context"
	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByID(ctx context.Context, teamID string) (*domain.Team, error)
	GetByName(ctx context.Context, name string) (*domain.Team, error)
}

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) TeamRepository {
	return &TeamRepo{pool: pool}
}

func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var teamID string
	err = tx.QueryRow(ctx,
		`INSERT INTO teams (name) VALUES ($1) RETURNING id`,
		team.TeamName,
	).Scan(&teamID)
	if err != nil {
		return err
	}

	// Обрабатываем каждого участника команды
	for _, member := range team.Members {
		var userID string

		// Пытаемся получить существующего пользователя
		err = tx.QueryRow(ctx,
			`SELECT id FROM users WHERE id = $1`,
			member.UserID,
		).Scan(&userID)

		if err != nil {
			// Пользователя нет в базе - добавляем его
			err = tx.QueryRow(ctx,
				`INSERT INTO users (id, username, is_active) VALUES ($1, $2, $3) RETURNING id`,
				member.UserID,
				member.Username,
				member.IsActive,
			).Scan(&userID)
			if err != nil {
				return err
			}
		}

		// Добавляем связь команда-пользователь
		_, err = tx.Exec(ctx,
			`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`,
			teamID,
			userID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamRepo) GetByID(ctx context.Context, teamID string) (*domain.Team, error) {
	var teamName string
	err := r.pool.QueryRow(ctx,
		`SELECT name FROM teams WHERE id = $1`,
		teamID,
	).Scan(&teamName)
	if err != nil {
		return nil, err
	}

	// Получаем всех участников команды
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.is_active 
         FROM users u
         JOIN team_members tm ON u.id = tm.user_id
         WHERE tm.team_id = $1`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (r *TeamRepo) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	var teamID string
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM teams WHERE name = $1`,
		name,
	).Scan(&teamID)
	if err != nil {
		return nil, err
	}

	// Получаем всех участников команды
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.is_active 
         FROM users u
         JOIN team_members tm ON u.id = tm.user_id
         WHERE tm.team_id = $1`,
		teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &domain.Team{
		TeamName: name,
		Members:  members,
	}, nil
}

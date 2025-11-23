package repository

import (
	"context"
	"github.com/Unitazavr/AvitoPR/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetPullRequests(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepository {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByUserID(ctx context.Context, userID string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT u.id, u.username, u.is_active, COALESCE(t.name, '') as team_name
		FROM users u
		LEFT JOIN team_members tm ON u.id = tm.user_id
		LEFT JOIN teams t ON tm.team_id = t.id
		WHERE u.id = $1
		LIMIT 1
	`, userID)

	var u domain.User
	if err := row.Scan(&u.UserID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
		return nil, domain.ErrNotFound
	}
	return &u, nil
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET is_active = $1 WHERE id = $2`, isActive, userID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.GetByUserID(ctx, userID)
}

func (r *UserRepo) GetPullRequests(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, pull_request_name, author_id, status FROM prs WHERE id IN (
    SELECT pr_id FROM pr_reviewers WHERE user_id = $1 )`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort

		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

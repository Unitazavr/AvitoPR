package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Unitazavr/AvitoPR/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrRepository interface {
	Create(ctx context.Context, pr *domain.PullRequestShort) error
	Merge(ctx context.Context, prId string) error
	Reassign(ctx context.Context, pullRequestId, oldUserId string) (newReviewerID string, err error)
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
}

type PrRepo struct {
	pool *pgxpool.Pool
}

func NewPrRepo(pool *pgxpool.Pool) PrRepository {
	return &PrRepo{pool: pool}
}

func (r *PrRepo) Create(ctx context.Context, pr *domain.PullRequestShort) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO prs (pull_request_name, author_id, status, created_at) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id`,
		pr.PullRequestName,
		pr.AuthorID,
		domain.PRStatusOpen,
		time.Now(),
	).Scan(&pr.PullRequestID)
	if err != nil {
		return err
	}

	// Получаем команду автора
	var teamID string
	err = tx.QueryRow(ctx,
		`SELECT team_id FROM team_members WHERE user_id = $1 LIMIT 1`,
		pr.AuthorID,
	).Scan(&teamID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Автор не состоит в команде - PR создается без ревьюверов
			return tx.Commit(ctx)
		}
		return err
	}

	// Получаем до двух активных участников команды, исключая автора
	rows, err := tx.Query(ctx,
		`SELECT u.id 
		 FROM users u
		 JOIN team_members tm ON u.id = tm.user_id
		 WHERE tm.team_id = $1 
		   AND u.id != $2 
		   AND u.is_active = true
		 ORDER BY RANDOM()
		 LIMIT 2`,
		teamID,
		pr.AuthorID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Назначаем ревьюверов
	for _, reviewerID := range reviewers {
		_, err = tx.Exec(ctx,
			`INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`,
			pr.PullRequestID,
			reviewerID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PrRepo) Merge(ctx context.Context, prId string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE prs 
		 SET status = $1, merged_at = $2 
		 WHERE id = $3 AND status = $4`,
		domain.PRStatusMerged,
		time.Now(),
		prId,
		domain.PRStatusOpen,
	)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("PR not found or already merged")
	}

	return nil
}

func (r *PrRepo) Reassign(ctx context.Context, pullRequestId, oldUserId string) (newReviewerID string, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// Проверяем, что PR не в статусе MERGED
	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM prs WHERE id = $1`,
		pullRequestId,
	).Scan(&status)
	if err != nil {
		return "", err
	}

	if status == string(domain.PRStatusMerged) {
		return "", fmt.Errorf("cannot reassign reviewers for merged PR")
	}

	// Проверяем, что oldUserId является ревьювером этого PR
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2)`,
		pullRequestId,
		oldUserId,
	).Scan(&exists)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("user is not a reviewer of this PR")
	}

	// Получаем команду заменяемого ревьювера
	var teamID string
	err = tx.QueryRow(ctx,
		`SELECT team_id FROM team_members WHERE user_id = $1 LIMIT 1`,
		oldUserId,
	).Scan(&teamID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user is not in any team")
		}
		return "", err
	}

	// Получаем автора PR
	var authorID string
	err = tx.QueryRow(ctx,
		`SELECT author_id FROM prs WHERE id = $1`,
		pullRequestId,
	).Scan(&authorID)
	if err != nil {
		return "", err
	}

	// Получаем текущих ревьюверов PR
	rows, err := tx.Query(ctx,
		`SELECT user_id FROM pr_reviewers WHERE pr_id = $1`,
		pullRequestId,
	)
	if err != nil {
		return "", err
	}

	var currentReviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			rows.Close()
			return "", err
		}
		currentReviewers = append(currentReviewers, reviewerID)
	}
	rows.Close()

	if err = rows.Err(); err != nil {
		return "", err
	}

	// Находим случайного активного участника из команды, исключая автора и текущих ревьюверов
	query := `SELECT u.id 
		 FROM users u
		 JOIN team_members tm ON u.id = tm.user_id
		 WHERE tm.team_id = $1 
		   AND u.id != $2 
		   AND u.is_active = true`

	args := []interface{}{teamID, authorID}
	for i, reviewerID := range currentReviewers {
		query += fmt.Sprintf(" AND u.id != $%d", i+3)
		args = append(args, reviewerID)
	}

	query += " ORDER BY RANDOM() LIMIT 1"

	err = tx.QueryRow(ctx, query, args...).Scan(&newReviewerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no available reviewers in team")
		}
		return "", err
	}

	// Удаляем старого ревьювера
	_, err = tx.Exec(ctx,
		`DELETE FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2`,
		pullRequestId,
		oldUserId,
	)
	if err != nil {
		return "", err
	}

	// Добавляем нового ревьювера
	_, err = tx.Exec(ctx,
		`INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`,
		pullRequestId,
		newReviewerID,
	)
	if err != nil {
		return "", err
	}

	return newReviewerID, tx.Commit(ctx)
}

func (r *PrRepo) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Получаем основные данные PR
	var pr domain.PullRequest
	err := r.pool.QueryRow(ctx,
		`SELECT id, pull_request_name, author_id, status, created_at, merged_at
		 FROM prs WHERE id = $1`,
		prID,
	).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		return nil, err
	}

	// Получаем список ревьюверов
	rows, err := r.pool.Query(ctx,
		`SELECT user_id FROM pr_reviewers WHERE pr_id = $1`,
		prID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

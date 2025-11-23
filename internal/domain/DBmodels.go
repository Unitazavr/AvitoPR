package domain

import (
	"github.com/google/uuid"
	"time"
)

type DBUser struct {
	ID       uuid.UUID `db:"id"`
	Username string    `db:"username"`
	IsActive bool      `db:"is_active"`
}

type DBTeam struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

type DBTeamMember struct {
	TeamID uuid.UUID `db:"team_id"`
	UserID uuid.UUID `db:"user_id"`
}

type DBPR struct {
	ID              uuid.UUID  `db:"id"`
	PullRequestName string     `db:"pull_request_name"`
	AuthorID        uuid.UUID  `db:"author_id"`
	Status          PRStatus   `db:"status"`
	CreatedAt       time.Time  `db:"created_at"`
	MergedAt        *time.Time `db:"merged_at"`
}

type DBPRReviewer struct {
	PRID   uuid.UUID `db:"pr_id"`
	UserID uuid.UUID `db:"user_id"`
}

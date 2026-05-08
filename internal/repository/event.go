package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"errors"
	"time"
)

type Event struct {
	ID          int64
	Title       string
	Description string
	Source      string
	OccurredAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *Event) error {
	if e == nil {
		return errors.New("event is nil")
	}

	query := `
		INSERT INTO events (title, description, source, occurred_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, e.Title, e.Description, e.Source, e.OccurredAt).Scan(&e.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id int64) (*Event, error) {
	const query = `
		SELECT id, title, description, source, occurred_at, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	var e Event
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.Title,
		&e.Description,
		&e.Source,
		&e.OccurredAt,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *EventRepository) List(ctx context.Context, limit, offset int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const query = `
		SELECT id, title, description, source, occurred_at, created_at, updated_at
		FROM events
		ORDER BY occurred_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Source,
			&e.OccurredAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

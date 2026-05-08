package service

import (
	"context"
	"errors"
	"time"

	"global-event-feed/internal/model"
	"global-event-feed/internal/repository"
)

// EventService handles business logic.
type EventService struct {
	repo *repository.EventRepository
}

// NewEventService creates a new service instance.
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// CreateEvent validates and inserts an event.
func (s *EventService) CreateEvent(ctx context.Context, e *model.Event) error {
	if e == nil {
		return errors.New("event is required")
	}

	// Basic validation.
	if e.Type == "" || e.Title == "" || e.Location == "" {
		return errors.New("type, title, and location are required")
	}

	// Ensure timestamp is set.
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	repoEvent := &repository.Event{
		Title:       e.Title,
		Description: e.Description,
		Source:      e.Location,
		OccurredAt:  e.Timestamp,
	}

	if err := s.repo.Create(ctx, repoEvent); err != nil {
		return err
	}

	e.ID = int(repoEvent.ID)
	return nil
}

// GetEvents returns latest events up to limit.
func (s *EventService) GetEvents(ctx context.Context, limit int) ([]model.Event, error) {
	if limit <= 0 {
		limit = 50
	}

	repoEvents, err := s.repo.List(ctx, limit, 0)
	if err != nil {
		return nil, err
	}

	events := make([]model.Event, 0, len(repoEvents))
	for _, re := range repoEvents {
		events = append(events, model.Event{
			ID:          int(re.ID),
			Type:        "", // repository schema has no dedicated type field
			Title:       re.Title,
			Description: re.Description,
			Location:    re.Source,
			Timestamp:   re.OccurredAt,
			CreatedAt:   re.CreatedAt,
		})
	}

	return events, nil
}

package service
import (
	"context"
	"errors"
	"time"

	"global-event-feed/internal/model"
	"global-event-feed/internal/repository"
)


// EventService han9dles business logic
type EventService struct {
	repo *repository.EventRepository
}


// NewEventService creates a new service instance
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}



// CreateEvent validates and inserts an event
func (s *EventService) CreateEvent(ctx context.Context, e *model.Event) error {
   // Basic Validation
   if e.Type == "" || e.Title == "" || e.Location == "" {
      return errors.New("type, title, and location are required")
   }


   // Ensure timestamp is inserts
   if e.Timestamp.isZero(){
    e.Timestamp = time.Now()
   }


   return s.repo.InsertEvent(ctx, e)
}


func (s *EventService) getEvents(ctx context.Context, limit int) ([]model.Event, error) {
	if limit <=0 {
		limit = 50
	}

	return s.repo.GetEvents(ctx, limit)
}

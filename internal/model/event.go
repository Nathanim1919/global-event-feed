package model

import "time"

// Event represents a single global event
type Event struct {
	ID          int       `json:"id"`          // auto-incremented in DB
	Type        string    `json:"type"`        // event type (update, alert, info)
	Title       string    `json:"title"`       // short title
	Description string    `json:"description"` // longer description
	Location    string    `json:"location"`    // e.g., "New York"
	Timestamp   time.Time `json:"timestamp"`   // when event happened
	CreatedAt   time.Time `json:"created_at"`  // when it was inserted
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"global-event-feed/internal/model"
	"global-event-feed/internal/service"

	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// POST /events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var e model.Event
	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.svc.CreateEvent(r.Context(), &e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

// GET /events?limit=50
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	events, err := h.svc.GetEvents(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}


// GET /events/{id}
func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	event, err := h.svc.GetEventByID(r.Context(), id)
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(event)
}

// RegisterRoutes attaches handlers to router
func (h *EventHandler) RegisterRoutes(r chi.Router) {
	r.Post("/events", h.CreateEvent)
	r.Get("/events", h.GetEvents)
	r.Get("/events/{id}", h.GetEventByID)
}

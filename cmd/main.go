package main

import (
	"log"
	"net/http"

	"global-event-feed/config"
	"global-event-feed/internal/handler"
	"global-event-feed/internal/repository"
	"global-event-feed/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect DB
	db := repository.ConnectDB(cfg)
	defer db.Close()

	// Initialize layers
	eventRepo := repository.NewEventRepository(db)
	eventSvc := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventSvc)

	// Router
	r := chi.NewRouter()
	eventHandler.RegisterRoutes(r)

	log.Printf("Server running on port %s", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}

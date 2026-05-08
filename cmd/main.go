package main

import (
	"log"
	"net/http"
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

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

	// log.Printf("Server running on port %s", cfg.Port)
	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: r,
	}


    // Run Server in Goroutine
    go func(){
      log.Println("Server started on :"+cfg.Port)

      if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
          log.Fatalf("server error: %v", err)
      }
    }()


    // Channel to receive OS signals
    stop := make(chan os.Signal, 1)


    // Listen for shutdown signals
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)


    // Block until signal received
    <-stop


    log.Println("Shutting down server ...")

    // Timeout for gracefull shutdown
    ctx, cancle := context.WithTimeout(
    	context.Background(),
        5*time.Second,
    )
    defer cancle()


    // Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}


	log.Println("Server exited cleanly")
}

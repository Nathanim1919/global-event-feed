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
	"global-event-feed/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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
	zapLogger, err := logger.NewLogger()
	 if err != nil {
	      log.Fatal("failed to initialize logger: ", err)
	  }
	defer zapLogger.Sync()

	// Router
	r := chi.NewRouter()

	// Run middlewares
		r.Use(func(next http.Handler) http.Handler {
	      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	          w.Header().Set("Content-Type", "application/json")
	          next.ServeHTTP(w, r)
	      })
	  })
	r.Use(middleware.Logger)   // logs every request (method, path, duration)
	r.Use(middleware.Recoverer)  // caches panics so the server doesnt crash

	eventHandler.RegisterRoutes(r)

	// log.Printf("Server running on port %s", cfg.Port)
	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: r,
	}


    // Run Server in Goroutine
    go func(){
    	zapLogger.Info("Server running at: ", zap.String("Port", cfg.Port))

	    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		    zapLogger.Fatal("server error", zap.Error(err))
	    }
    }()


    // Channel to receive OS signals
    stop := make(chan os.Signal, 1)


    // Listen for shutdown signals
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)


    // Block until signal received
    <-stop


    zapLogger.Info("Shutting down server ...")

    // Timeout for gracefull shutdown
    ctx, cancle := context.WithTimeout(
    	context.Background(),
        5*time.Second,
    )
    defer cancle()


    // Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("shutdown failed: %v", zap.Error(err))
	}


	zapLogger.Info("Server exited cleanly")
}

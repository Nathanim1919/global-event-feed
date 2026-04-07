package main

import (
	"log"
	"net/http"
	"global-event-feed/config"
	"global-event-feed/internal/repository"
)


func main(){
	// Load environment config
	cfg := config.LoadConfig()

	// connect to Postgres
	dbPool := repository.ConnectDB(config)
	defer dbPool.Close() // close pool when app exists

	// TODO: initlaize router and attach handlers


	log.Printf("Server ready on port %s\n", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, nil) // Placeholder for now
}

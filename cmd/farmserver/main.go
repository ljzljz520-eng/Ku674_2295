package main

import (
	"flag"
	"log"
	"net/http"

	"farm-sensor-platform/internal/api"
	"farm-sensor-platform/internal/config"
	"farm-sensor-platform/internal/observability"
	"farm-sensor-platform/internal/persistence"
	"farm-sensor-platform/internal/sensors"
)

func main() {
	listen := flag.String("listen", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "farm-sensors.db", "bbolt database path")
	flag.Parse()
	settings := config.Default()
	settings.ListenAddress = *listen
	settings.DatabasePath = *dbPath
	store, err := persistence.Open(settings.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	logger := observability.NewLogger()
	service := sensors.NewService(store, logger, settings)
	handler := api.NewHandler(service, logger)
	logger.Info("farm sensor service listening", "address", settings.ListenAddress)
	if err := http.ListenAndServe(settings.ListenAddress, handler); err != nil {
		logger.Error("server stopped", "error", err)
	}
}

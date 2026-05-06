package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type config struct {
	HTTPAddr          string
	DatabaseDSN       string
	SchedulerCronSpec string
	SchedulerLocation *time.Location
}

func loadConfig() (config, error) {
	locationName := envOrDefault("SCHEDULER_TIMEZONE", "UTC")
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return config{}, fmt.Errorf("load scheduler timezone: %w", err)
	}

	cfg := config{
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:       envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		SchedulerCronSpec: envOrDefault("SCHEDULER_CRON_SPEC", "5 0 * * *"),
		SchedulerLocation: location,
	}

	if cfg.DatabaseDSN == "" {
		return config{}, errors.New("DATABASE_DSN is required")
	}
	if cfg.SchedulerCronSpec == "" {
		return config{}, errors.New("SCHEDULER_CRON_SPEC is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

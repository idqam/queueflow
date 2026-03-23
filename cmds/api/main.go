package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"owen/queueflow/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("[NON-BLOCK ERROR] .env not loaded (%v), continuing with existing environment", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cfgJson, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[NON-BLOCK ERROR] failed to marshal config: %v", err)
	} else {
		log.Printf("starting API with config:\n%s", string(cfgJson))
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: NewRouter(),
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

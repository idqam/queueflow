package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

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


	pingReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	pingResp := httptest.NewRecorder()
	Ping(pingResp, pingReq)
	if pingResp.Code != http.StatusOK || pingResp.Body.String() != "pong" {
		log.Fatalf("ping health check failed: code=%d body=%q", pingResp.Code, pingResp.Body.String())
	}
	log.Printf("ping endpoint check passed")



	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: NewRouter(),
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

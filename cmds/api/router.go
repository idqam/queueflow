package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", HealthCheck)
	r.Get("/ping", Ping)
	return r
}

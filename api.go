package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type API struct {
	engine *FileEngine
}

func NewAPI(engine *FileEngine) *API {
	return &API{engine: engine}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /events", a.handlePostEvent)
	mux.HandleFunc("GET /events/{id}", a.handleGetEvent)
	mux.HandleFunc("GET /stats", a.handleGetStats)
}

func (a *API) handlePostEvent(w http.ResponseWriter, r *http.Request) {
	var event map[string]any

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()

	event["id"] = id
	event["createdAt"] = time.Now().UTC()

	data, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "failed to marshal event", http.StatusInternalServerError)
		return
	}

	line := string(data)

	if err := a.engine.Set(id, line); err != nil {
		http.Error(w, "failed to store event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write(data)
}

func (a *API) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	line, err := a.engine.Get(id)
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, _ = w.Write([]byte(line))
}

func (a *API) handleGetStats(w http.ResponseWriter, r *http.Request) {
	total, bytes, err := a.engine.Stats()
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"total": total,
		"bytes": bytes,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

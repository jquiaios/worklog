package server

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
)

//go:embed ui/index.html
var staticFiles embed.FS

type Server struct {
	store *db.DB
}

func New(store *db.DB) http.Handler {
	s := &Server{store: store}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /api/entries", s.handleListEntries)
	mux.HandleFunc("POST /api/entries", s.handleCreateEntry)
	mux.HandleFunc("PUT /api/entries/{id}", s.handleUpdateEntry)
	mux.HandleFunc("DELETE /api/entries/{id}", s.handleDeleteEntry)

	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("ui/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		log.Printf("server: writing index: %v", err)
	}
}

type entryResponse struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

func toResponse(e entry.Entry) entryResponse {
	return entryResponse{
		ID:        e.ID,
		Type:      string(e.Type),
		Body:      e.Body,
		CreatedAt: e.CreatedAt.Local().Format("2006-01-02"),
	}
}

func (s *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	entryType := q.Get("type")

	var from, to time.Time
	if quarterStr := q.Get("quarter"); quarterStr != "" {
		quarter, _ := strconv.Atoi(quarterStr)
		year, _ := strconv.Atoi(q.Get("year"))
		if quarter >= 1 && quarter <= 4 && year > 0 {
			month := time.Month((quarter-1)*3 + 1)
			from = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
			to = from.AddDate(0, 3, 0)
		}
	}

	entries, err := s.store.List(entryType, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]entryResponse, len(entries))
	for i, e := range entries {
		resp[i] = toResponse(e)
	}
	writeJSON(w, resp)
}

type createRequest struct {
	Type string `json:"type"`
	Body string `json:"body"`
}

func (s *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Body == "" {
		http.Error(w, "body required", http.StatusBadRequest)
		return
	}
	if req.Type != string(entry.Highlight) && req.Type != string(entry.Lowlight) {
		http.Error(w, "type must be highlight or lowlight", http.StatusBadRequest)
		return
	}

	e := entry.Entry{
		Type:      entry.Type(req.Type),
		Body:      req.Body,
		CreatedAt: time.Now(),
	}
	id, err := s.store.Insert(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.ID = id
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, toResponse(e))
}

type updateRequest struct {
	Body string `json:"body"`
}

func (s *Server) handleUpdateEntry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Body == "" {
		http.Error(w, "body required", http.StatusBadRequest)
		return
	}
	if err := s.store.Update(id, req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := s.store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("server: encoding JSON response: %v", err)
	}
}

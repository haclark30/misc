package server

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"

	"misc/cmd/web"
	"misc/internal/models"

	"github.com/a-h/templ"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("/health", s.healthHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	mux.Handle("/assets/", fileServer)
	mux.Handle("/web", templ.Handler(web.HelloForm()))
	mux.HandleFunc("/hello", web.HelloWebHandler)
	mux.HandleFunc("POST /habiticaEvent", s.HabiticaWebhookHandler)
	mux.HandleFunc("POST /todoistEvent", s.TodoistWebhookHandler)
	mux.HandleFunc("GET /widget", s.WidgetHandler)

	return mux
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	req, _ := io.ReadAll(r.Body)
	slog.Info("got req", "method", r.Method, "body", string(req))
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) HabiticaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var req models.HabiticaWebhook
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("error decoding request", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Task.Type != models.HabiticaHabitType {
		slog.Info("got non habit task", "event", req)
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("checking habit", "id", req.Task.Id, "name", req.Task.Text)
	err = s.habService.CheckMinHabit(req.Task.Id, req.Task.Up)
	if err != nil {
		slog.Error("error checking habit", "err", err)
	}
}

func (s *Server) TodoistWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var req models.TodoistWebhook
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("error decoding request", "err", err)
		return
	}
	slog.Info("got todoist event", "taskName", req.EventData.Content)

	err = s.todoHabService.ScoreTask(req.EventData.Content, req.EventData.ProjectId)

	if err != nil {
		slog.Error("error scoring task", "err", err)
		return
	}
}

func (s *Server) WidgetHandler(w http.ResponseWriter, r *http.Request) {
	resp := s.widgetService.GetWidgetResponse()
	if len(resp.Errors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"misc/clients/habitica"
	"misc/clients/todoist"
	"misc/internal/database"
	"misc/internal/models"
	"misc/internal/services"
)

type WidgetService interface {
	GetWidgetResponse() models.WidgetResponse
}
type Server struct {
	port int

	db             database.Service
	habService     services.HabiticaMinHabitService
	todoHabService services.TodoistHabiticaService
	widgetService  WidgetService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}
	habClient := habitica.NewHabiticaClient(
		os.Getenv("HABITICA_API_USER"),
		os.Getenv("HABITICA_API_KEY"),
	)

	todoistRestClient := todoist.NewClient(os.Getenv("TODOIST_API_KEY"))
	todoistSyncClient := todoist.NewSyncClient(os.Getenv("TODOIST_API_KEY"))

	todoistService := services.NewTodoistService(todoistRestClient, todoistSyncClient)
	NewServer.habService = services.NewHabitcaMinHabitService(
		NewServer.db,
		&habClient,
	)

	NewServer.todoHabService = services.NewTodoistHabiticaService(
		NewServer.db,
		&habClient,
	)

	NewServer.widgetService = services.NewWidgetService(&habClient, &todoistService)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

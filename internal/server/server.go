package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"misc/clients"
	"misc/internal/database"
	"misc/internal/services"
)

type Server struct {
	port int

	db             database.Service
	habService     services.HabiticaMinHabitService
	todoHabService services.TodoistHabiticaService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}
	habClient := clients.NewHabiticaClient(
		os.Getenv("HABITICA_API_USER"),
		os.Getenv("HABITICA_API_KEY"),
	)

	NewServer.habService = *services.NewHabitcaMinHabitService(
		NewServer.db,
		&habClient,
	)

	NewServer.todoHabService = services.NewTodoistHabiticaService(
		NewServer.db,
		&habClient,
	)

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

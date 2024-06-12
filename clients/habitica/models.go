package habitica

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// generic struct for habitica responses, compose into different types
type HabiticaResponse struct {
	Success bool `json:"success"`
}

type HabiticaError struct {
	Response  *http.Response `json:"-"`
	ErrorCode string         `json:"error"`
	Message   string         `json:"message"`
}

func (h HabiticaError) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		h.Response.Request.Method, h.Response.Request.URL,
		h.Response.StatusCode, h.Message)
}

// all tasks have these, use composition on different task types
type Task struct {
	ID     string   `json:"id"`
	UserID string   `json:"userId"`
	Text   string   `json:"text"`
	Type   string   `json:"type"`
	Notes  string   `json:"notes"`
	Tags   []string `json:"tags"`
}

type Habit struct {
	Up          bool   `json:"up"`
	Down        bool   `json:"down"`
	CounterUp   int    `json:"counterUp"`
	CounterDown int    `json:"counterDown"`
	Frequency   string `json:"frequency"`
	Task
}

type Daily struct {
	Completed bool   `json:"completed"`
	Repeat    Repeat `json:"repeat"`
	IsDue     bool   `json:"isDue"`
	Task
}

type Repeat struct {
	Mon bool `json:"m"`
	Tue bool `json:"t"`
	Wed bool `json:"w"`
	Thu bool `json:"th"`
	Fri bool `json:"f"`
	Sat bool `json:"s"`
	Sun bool `json:"su"`
}

type HabitsResponse struct {
	Data []Habit `json:"data"`
	HabiticaResponse
}

type DailysResponse struct {
	Data []Daily `json:"data"`
	HabiticaResponse
}

func (h *Habit) ParseGoal() (int, error) {
	return strconv.Atoi(strings.Trim(h.Notes, "Goal: "))
}

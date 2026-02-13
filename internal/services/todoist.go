package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"misc/clients/todoist"
	"net/http"

	"github.com/google/go-querystring/query"
)

type TodoistService struct {
	restClient *todoist.TodoistRestClient
	syncClient *todoist.TodoistSyncClient
}

func NewTodoistService(restClient *todoist.TodoistRestClient, syncClient *todoist.TodoistSyncClient) TodoistService {
	return TodoistService{restClient, syncClient}
}

func (t *TodoistService) GetTasks(filter *todoist.TaskFilterOptions) ([]todoist.Task, error) {
	req, err := t.restClient.NewTodoistRequest(http.MethodGet, "tasks/filter", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create todoist tasks request: %w", err)
	}

	if filter != nil {
		v, err := query.Values(filter)
		if err != nil {
			return nil, fmt.Errorf("unable to encode filter: %w", err)
		}
		req.URL.RawQuery = v.Encode()
	}

	resp, err := t.restClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling todoist for tasks: %w", err)
	}

	var todoResp todoist.TaskResp

	err = json.NewDecoder(resp.Body).Decode(&todoResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding todoist task resp: %w", err)
	}

	return todoResp.Tasks, nil
}

func (t *TodoistService) GetStats() (todoist.Stats, error) {
	req, err := t.restClient.NewTodoistRequest(http.MethodGet, "tasks/completed/stats", nil)
	if err != nil {
		return todoist.Stats{}, fmt.Errorf("unable to create stats req: %w", err)
	}

	resp, err := t.restClient.Client.Do(req)
	if err != nil {
		return todoist.Stats{}, fmt.Errorf("error calling todoist for stats: %w", err)
	}

	var stats todoist.Stats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return todoist.Stats{}, fmt.Errorf("error decoding todoist stats resp: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return todoist.Stats{}, errors.New("error from todoist api")
	}

	return stats, nil
}

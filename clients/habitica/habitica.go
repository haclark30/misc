package habitica

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type HabiticaClient struct {
	apiUser string
	apiKey  string
	client  *http.Client
}

const habUrl = "https://habitica.com/api/v3"

func NewHabiticaClient(apiUser, apiKey string) HabiticaClient {
	return HabiticaClient{apiUser: apiUser, apiKey: apiKey, client: http.DefaultClient}
}

func (h *HabiticaClient) habiticaRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-api-user", h.apiUser)
	req.Header.Add("x-api-key", h.apiKey)
	req.Header.Add("x-client", fmt.Sprintf("%s-misc-webhooks", h.apiUser))
	return req, nil
}

func (h *HabiticaClient) ScoreDaily(dailyId string) error {
	req, err := h.habiticaRequest(
		http.MethodPost,
		fmt.Sprintf("%s/tasks/%s/score/up", habUrl, dailyId),
		nil,
	)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("error calling habitica api", "code", resp.StatusCode, "resp", respBody, "url", req.URL)
		return fmt.Errorf("got non-200 status code: %d", resp.StatusCode)
	}
	return nil
}

func (h *HabiticaClient) GetHabits() ([]Habit, error) {
	req, err := h.habiticaRequest(
		http.MethodGet,
		fmt.Sprintf("%s/tasks/user", habUrl),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("type", "habits")
	req.URL.RawQuery = q.Encode()

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request: %w", err)
	}
	defer resp.Body.Close()

	var habResp HabitsResponse
	err = json.NewDecoder(resp.Body).Decode(&habResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return habResp.Data, nil
}

func (h *HabiticaClient) GetDailys() ([]Daily, error) {
	req, err := h.habiticaRequest(
		http.MethodGet,
		fmt.Sprintf("%s/tasks/user", habUrl),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("type", "dailys")
	req.URL.RawQuery = q.Encode()

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request: %w", err)
	}
	defer resp.Body.Close()

	var habResp DailysResponse
	err = json.NewDecoder(resp.Body).Decode(&habResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return habResp.Data, nil
}

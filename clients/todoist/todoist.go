package todoist

import (
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.todoist.com/api/v1"
const baseSyncUrl = "https://api.todoist.com/sync/v9"

type TodoistClient struct {
	apiKey  string
	Client  *http.Client
	BaseUrl string
}

type APIError struct {
	Response     *http.Response `json:"-"`
	ErrorMessage string         `json:"error"`
	ErrorCode    int            `json:"error_code"`
	HttpCode     int            `json:"http_code"`
}

func (err APIError) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		err.Response.Request.Method, err.Response.Request.URL,
		err.Response.StatusCode, err.ErrorMessage)
}

type TodoistRestClient struct {
	TodoistClient
}

type TodoistSyncClient struct {
	TodoistClient
}

func NewClient(apiKey string) *TodoistRestClient {
	t := &TodoistClient{
		apiKey:  apiKey,
		Client:  http.DefaultClient,
		BaseUrl: baseURL,
	}
	c := &TodoistRestClient{
		TodoistClient: *t,
	}

	return c
}

func NewSyncClient(apiKey string) *TodoistSyncClient {
	t := &TodoistClient{
		apiKey:  apiKey,
		Client:  http.DefaultClient,
		BaseUrl: baseSyncUrl,
	}
	c := &TodoistSyncClient{
		TodoistClient: *t,
	}

	return c
}

func (c *TodoistClient) NewTodoistRequest(method, urlPath string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", c.BaseUrl, urlPath)
	authHeader := fmt.Sprintf("Bearer %s", c.apiKey)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)

	return req, nil
}

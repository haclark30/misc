package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

type ActivityResponse struct {
	Summary FitbitSummary `json:"summary"`
	Goals   FitbitGoals   `json:"goals"`
}

type WeightResponse struct {
	WeightRecords []WeightRecord `json:"weight"`
}

type WeightRecord struct {
	Weight float64 `json:"weight"`
	Date   string  `json:"date"`
}

type FitbitSummary struct {
	Steps             int `json:"steps"`
	VeryActiveMinutes int `json:"veryActiveMinutes"`
}

type FitbitGoals struct {
	ActiveMinutes int `json:"activeMinutes"`
	Steps         int `json:"steps"`
}

const expiryFmt = "2006-01-02T15:04:05Z07:00"

func createFitbitClient() *http.Client {

	codeChan := make(chan string)
	srv := http.NewServeMux()
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		urlVals := r.URL.Query()
		codeChan <- urlVals.Get("code")
	})
	redirectHost := os.Getenv("FITBIT_REDIRECT_HOST")
	if redirectHost == "" {
		redirectHost = "localhost"
	}
	var redirectUrl string
	if redirectHost == "localhost" {
		redirectUrl = "http://localhost:8001"
	} else {
		redirectUrl = fmt.Sprintf("https://%s", redirectHost)
	}
	go http.ListenAndServe(":8001", srv)

	conf := &oauth2.Config{
		ClientID:     os.Getenv("FITBIT_CLIENT_ID"),
		ClientSecret: os.Getenv("FITBIT_CLIENT_SECRET"),
		Scopes:       []string{"activity", "profile", "sleep", "nutrition", "weight", "sleep"},
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.fitbit.com/oauth2/authorize",
			TokenURL: "https://api.fitbit.com/oauth2/token",
		},
	}
	ctx := context.Background()
	token, err := loadToken()

	if err != nil || token == nil {
		verifier := oauth2.GenerateVerifier()

		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
		fmt.Printf("visit url for auth: %v\n", url)
		authToken := <-codeChan
		token, err = conf.Exchange(ctx, authToken, oauth2.VerifierOption(verifier))
		if err != nil {
			slog.Error("error getting token", "err", err)
		}
		saveToken(token)
	}

	if !token.Valid() {
		token, err = conf.TokenSource(ctx, token).Token()
		if err != nil {
			log.Fatal(err)
		}
		saveToken(token)
	}

	fitbitClient := conf.Client(ctx, token)
	return fitbitClient
}

func loadToken() (*oauth2.Token, error) {
	file, err := os.Open("./.fitbit_token/token.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	tokenExpiry, err := time.Parse(expiryFmt, token.Expiry)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       tokenExpiry,
	}, nil
}

func saveToken(token *oauth2.Token) {

	file, err := os.Create("./.fitbit_token/token.json")
	if err != nil {
		slog.Error("error creating fitbit token file")
	}

	defer file.Close()

	t := Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry.Format(expiryFmt),
	}

	json.NewEncoder(file).Encode(t)

	if err != nil {
		slog.Error("error encoding fitbit token")
	}
}

func getFitbitActivity(fitbitClient *http.Client) (ActivityResponse, error) {
	today := time.Now().Format("2006-01-02")
	resp, err := fitbitClient.Get(
		fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/date/%s.json", today),
	)
	if err != nil {
		slog.Error("error making request to fitbit", "err", err)
		return ActivityResponse{}, fmt.Errorf("error making request to fitbit: %w", err)
	}

	defer resp.Body.Close()

	slog.Info("got resp code", "code", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("error code calling fitbit", "statusCode", resp.StatusCode, "respBody", body)
		return ActivityResponse{}, fmt.Errorf("error code calling fitbit: %d", resp.StatusCode)
	}

	act := ActivityResponse{}
	err = json.NewDecoder(resp.Body).Decode(&act)
	if err != nil {
		slog.Error("error decoding fitbit response", "err", err)
		return ActivityResponse{}, fmt.Errorf("error decoding fitbit response: %w", err)
	}
	slog.Info("got activity", "act", act)
	return act, nil
}

func getFitbitWeight(fitbitClient *http.Client) (WeightResponse, error) {
	today := time.Now().Format("2006-01-02")
	var weightResponse WeightResponse

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.fitbit.com/1/user/-/body/log/weight/date/%s/7d.json", today),
		nil,
	)
	if err != nil {
		return weightResponse, err
	}

	req.Header.Add("accept-language", "en_US")

	resp, err := fitbitClient.Do(req)
	if err != nil {
		return weightResponse, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&weightResponse)
	if err != nil {
		return weightResponse, err
	}

	return weightResponse, err
}

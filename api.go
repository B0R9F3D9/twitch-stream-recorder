package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type TwitchAPI struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewTwitchAPI(clientID, clientSecret string) *TwitchAPI {
	return &TwitchAPI{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TwitchAPI) fetchAccessToken() (string, error) {
	baseURL := "https://id.twitch.tv/oauth2/token"
	params := url.Values{}
	params.Add("client_id", t.clientID)
	params.Add("client_secret", t.clientSecret)
	params.Add("grant_type", "client_credentials")
	url := baseURL + "?" + params.Encode()

	resp, err := t.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf(
				"unexpected status code %d, body read error: %w",
				resp.StatusCode,
				err,
			)
		}
		return "", fmt.Errorf(
			"unexpected status code %d, body: %s",
			resp.StatusCode,
			body,
		)
	}

	data := make(map[string]any)
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("invalid or missing token, response: %v", data)
	}
	return token, nil
}

func (t *TwitchAPI) isStreamOnline(username, accessToken string) (bool, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.twitch.tv/helix/streams?user_login="+username,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Client-ID", t.clientID)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf(
				"unexpected status code %d, body read error: %w",
				resp.StatusCode,
				err,
			)
		}
		return false, fmt.Errorf(
			"unexpected status code %d, body: %s",
			resp.StatusCode,
			body,
		)
	}

	data := make(map[string]any)
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	streams, ok := data["data"].([]any)
	if !ok {
		return false, fmt.Errorf("invalid or missing data, response: %v", data)
	}
	return len(streams) > 0, nil
}

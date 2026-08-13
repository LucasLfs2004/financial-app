package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lucas/financial-api/internal/platform/config"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnavailable  = errors.New("authentication service unavailable")
)

type Principal struct {
	UserID string
	Email  string
}

type Client struct {
	userEndpoint   string
	publishableKey string
	httpClient     *http.Client
}

func NewClient(cfg config.SupabaseConfig) *Client {
	return &Client{
		userEndpoint:   strings.TrimRight(cfg.URL, "/") + "/auth/v1/user",
		publishableKey: cfg.PublishableKey,
		httpClient: &http.Client{
			Timeout: cfg.AuthTimeout,
		},
	}
}

func (client *Client) Authenticate(ctx context.Context, accessToken string) (Principal, error) {
	if strings.TrimSpace(accessToken) == "" {
		return Principal{}, ErrUnauthorized
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.userEndpoint, nil)
	if err != nil {
		return Principal{}, fmt.Errorf("create authentication request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("apikey", client.publishableKey)

	response, err := client.httpClient.Do(request)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()

	switch {
	case response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden:
		return Principal{}, ErrUnauthorized
	case response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices:
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return Principal{}, fmt.Errorf("%w: auth returned status %d", ErrUnavailable, response.StatusCode)
	}

	var payload struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return Principal{}, fmt.Errorf("%w: decode user: %v", ErrUnavailable, err)
	}
	if strings.TrimSpace(payload.ID) == "" {
		return Principal{}, fmt.Errorf("%w: auth response has no user id", ErrUnavailable)
	}

	return Principal{
		UserID: payload.ID,
		Email:  payload.Email,
	}, nil
}

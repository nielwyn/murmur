package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nielwyn/murmur/internal/auth"

	"github.com/google/uuid"
)

// apiClient is a thin HTTP client for the same API the web frontend uses.
type apiClient struct {
	baseURL string
	http    *http.Client
	token   string
}

func newAPIClient(baseURL, token string) *apiClient {
	return &apiClient{baseURL: baseURL, http: &http.Client{}, token: token}
}

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }

type userResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type feedResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	CreatorName string    `json:"creator_name,omitempty"`
}

type followResponse struct {
	FeedID   uuid.UUID `json:"feed_id"`
	FeedName string    `json:"feed_name,omitempty"`
	FeedURL  string    `json:"feed_url,omitempty"`
}

// do sends a request and decodes the JSON response into out (if non-nil).
// It also returns the session token, if the response set one.
func (c *apiClient) do(ctx context.Context, method, path string, body, out any) (string, error) {
	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: c.token})
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach the API at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	var token string
	for _, ck := range resp.Cookies() {
		if ck.Name == auth.CookieName {
			token = ck.Value
		}
	}

	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		if e.Error == "" {
			e.Error = resp.Status
		}
		return token, &apiError{status: resp.StatusCode, message: e.Error}
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return token, fmt.Errorf("decoding response: %w", err)
		}
	}

	return token, nil
}

func (c *apiClient) Register(ctx context.Context, username, email, password string) (userResponse, string, error) {
	var u userResponse
	token, err := c.do(ctx, http.MethodPost, "/api/register", map[string]string{
		"username": username, "email": email, "password": password,
	}, &u)
	return u, token, err
}

func (c *apiClient) Login(ctx context.Context, username, password string) (userResponse, string, error) {
	var u userResponse
	token, err := c.do(ctx, http.MethodPost, "/api/login", map[string]string{
		"username": username, "password": password,
	}, &u)
	return u, token, err
}

func (c *apiClient) ListFeeds(ctx context.Context) ([]feedResponse, error) {
	var feeds []feedResponse
	_, err := c.do(ctx, http.MethodGet, "/api/feeds", nil, &feeds)
	return feeds, err
}

func (c *apiClient) CreateFeed(ctx context.Context, name, url string) (feedResponse, error) {
	var f feedResponse
	_, err := c.do(ctx, http.MethodPost, "/api/feeds", map[string]string{"name": name, "url": url}, &f)
	return f, err
}

func (c *apiClient) ListFollowing(ctx context.Context) ([]followResponse, error) {
	var follows []followResponse
	_, err := c.do(ctx, http.MethodGet, "/api/feeds/following", nil, &follows)
	return follows, err
}

func (c *apiClient) FollowFeed(ctx context.Context, feedID uuid.UUID) (followResponse, error) {
	var f followResponse
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/feeds/%s/follow", feedID), nil, &f)
	return f, err
}

func (c *apiClient) UnfollowFeed(ctx context.Context, feedID uuid.UUID) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/feeds/%s/follow", feedID), nil, nil)
	return err
}

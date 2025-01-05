package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type bearerAuthTransport struct {
	inner http.RoundTripper
	token string
}

func (b *bearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return b.inner.RoundTrip(req)
}

type Client struct {
	URL      string
	client   *http.Client
	identity string
}

func New(url string) *Client {
	return &Client{
		URL:    url,
		client: http.DefaultClient,
	}
}
func (r *Client) Identity() string {
	return r.identity
}

func (r *Client) Client() *http.Client {
	return r.client
}

func (r *Client) fetch(method, path string, body any, v any) error {
	marshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, r.URL+path, bytes.NewBuffer(marshaled))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
	}

	if v != nil {
		if err := json.NewDecoder(res.Body).Decode(v); err != nil {
			return err
		}
	}
	return nil
}

type AuthRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Token string `json:"token"`
}

func (r *Client) Authenticate(email, password string) error {
	var res AuthResponse
	err := r.fetch(
		"POST",
		"/api/collections/users/auth-with-password", AuthRequest{
			Identity: email,
			Password: password,
		},
		&res,
	)
	if err != nil {
		return err
	}

	r.client = &http.Client{
		Transport: &bearerAuthTransport{
			inner: http.DefaultTransport,
			token: res.Token,
		},
	}

	r.identity = email

	return nil
}

type Game struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Executable string `json:"executable"`
}

func (r *Client) GetGame(id string) (*Game, error) {
	var res Game
	err := r.fetch("GET", "/api/collections/games/records/"+id, nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

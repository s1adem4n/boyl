package igdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const TwitchTokenURL = "https://id.twitch.tv/oauth2/token"

type ClientCredentialsTokenSource struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	mu           sync.Mutex
	token        *oauth2.Token
}

func NewClientCredentialsTokenSource(clientID, clientSecret string) *ClientCredentialsTokenSource {
	return &ClientCredentialsTokenSource{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   http.DefaultClient,
	}
}

func (ts *ClientCredentialsTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.token.Valid() {
		return ts.token, nil
	}

	form := url.Values{}
	form.Set("client_id", ts.clientID)
	form.Set("client_secret", ts.clientSecret)
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", TwitchTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code for token: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	ts.token = &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return ts.token, nil
}

type ClientCredentialsRoundTripper struct {
	tokenSource oauth2.TokenSource
	clientID    string
}

func (rt *ClientCredentialsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := rt.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", rt.clientID)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	return http.DefaultTransport.RoundTrip(req)
}

func NewClientCredentialsClient(ctx context.Context, clientID, clientSecret string) *http.Client {
	ts := NewClientCredentialsTokenSource(clientID, clientSecret)
	rt := &ClientCredentialsRoundTripper{
		tokenSource: ts,
		clientID:    clientID,
	}
	return &http.Client{
		Transport: rt,
	}
}

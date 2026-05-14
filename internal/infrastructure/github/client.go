package githubclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GithubUser struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type GithubTeam struct {
	Org  string `json:"org"`
	Name string `json:"name"`
}

type GitHubClient interface {
	AuthorizeURL() string
	ExchangeCode(code string, codeVerifier string) (string, error)
	GetUser(token string) (GithubUser, error)
	GetUserOrgs(token string) ([]string, error)
	GetUserTeams(token string) ([]GithubTeam, error)
}

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	redirectURL  string
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

var (
	oauthAuthorizeURL = "https://github.com/login/oauth/authorize"
	oauthTokenURL     = "https://github.com/login/oauth/access_token"
	apiBaseURL        = "https://api.github.com"
)

func NewClient(clientID, clientSecret, redirectURL string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		redirectURL:  strings.TrimSpace(redirectURL),
	}
}

func (c *Client) AuthorizeURL() string {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("redirect_uri", c.redirectURL)
	values.Set("scope", "read:org")
	values.Set("allow_signup", "false")
	return fmt.Sprintf("%s?%s", oauthAuthorizeURL, values.Encode())
}

func (c *Client) ExchangeCode(code string, codeVerifier string) (string, error) {
	if code == "" {
		return "", errors.New("code is required")
	}

	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", c.redirectURL)

	if codeVerifier != "" {
		values.Set("code_verifier", codeVerifier)
		log.Printf("[GitHub] Token exchange request (with PKCE): code=%s (first 10 chars)", code[:min(10, len(code))])
	} else {
		log.Printf("[GitHub] Token exchange request (without PKCE): code=%s (first 10 chars)", code[:min(10, len(code))])
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, oauthTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[GitHub] HTTP request failed: %v", err)
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	log.Printf("[GitHub] Token response status: %d", resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("[GitHub] Token response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token exchange failed: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var tokenResp accessTokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Error != "" {
		msg := fmt.Sprintf("github error: %s", tokenResp.Error)
		if tokenResp.ErrorDesc != "" {
			msg = fmt.Sprintf("%s (%s)", msg, tokenResp.ErrorDesc)
		}
		log.Printf("[GitHub] Error response: %s", msg)
		return "", errors.New(msg)
	}

	log.Printf("[GitHub] Parsed response: AccessToken=%s, TokenType=%s, Scope=%s",
		tokenResp.AccessToken, tokenResp.TokenType, tokenResp.Scope)

	if tokenResp.AccessToken == "" {
		return "", errors.New("empty access token in GitHub response")
	}

	return tokenResp.AccessToken, nil
}

func (c *Client) GetUser(token string) (GithubUser, error) {
	if token == "" {
		return GithubUser{}, errors.New("token is required")
	}
	return c.fetchUser(token)
}

func (c *Client) GetUserOrgs(token string) ([]string, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	return c.fetchOrgs(token)
}

func (c *Client) GetUserTeams(token string) ([]GithubTeam, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	return c.fetchTeams(token)
}

func (c *Client) fetchUser(token string) (GithubUser, error) {
	var user GithubUser
	if err := c.doGet(token, apiBaseURL+"/user", &user); err != nil {
		return GithubUser{}, err
	}
	return user, nil
}

func (c *Client) fetchOrgs(token string) ([]string, error) {
	var orgs []struct {
		Login string `json:"login"`
	}
	if err := c.doGet(token, apiBaseURL+"/user/orgs", &orgs); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(orgs))
	for _, org := range orgs {
		result = append(result, org.Login)
	}
	return result, nil
}

func (c *Client) fetchTeams(token string) ([]GithubTeam, error) {
	var teams []struct {
		Name         string `json:"name"`
		Organization struct {
			Login string `json:"login"`
		} `json:"organization"`
	}
	if err := c.doGet(token, apiBaseURL+"/user/teams", &teams); err != nil {
		return nil, err
	}

	result := make([]GithubTeam, 0, len(teams))
	for _, team := range teams {
		result = append(result, GithubTeam{Org: team.Organization.Login, Name: team.Name})
	}
	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) doGet(token, urlStr string, target interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(token)))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "quickspaces-control-plane")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api request failed: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

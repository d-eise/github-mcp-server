// Package github provides a client for interacting with the GitHub API
// as part of the github-mcp-server MCP tool implementation.
package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with additional configuration
// and helper methods used by MCP tool handlers.
type Client struct {
	github *github.Client
	token  string
}

// NewClient creates a new GitHub API client using the provided personal
// access token. If token is empty, it falls back to the GITHUB_TOKEN
// environment variable. The token can be a classic PAT or a fine-grained
// personal access token.
//
// Note: fine-grained tokens may have restricted scopes; some API endpoints
// (e.g. listing all repos across orgs) may return 403 depending on permissions.
//
// Personal note: I prefer fine-grained tokens scoped to specific repos for
// better security hygiene, even if it means some endpoints are unavailable.
func NewClient(token string) (*Client, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required: set GITHUB_TOKEN environment variable or pass a token")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(httpClient)

	return &Client{
		github: ghClient,
		token:  token,
	}, nil
}

// NewClientWithHTTP creates a new GitHub API client using a custom HTTP client.
// This is primarily useful for testing with mock transports.
func NewClientWithHTTP(token string, httpClient *http.Client) *Client {
	ghClient := github.NewClient(httpClient)
	return &Client{
		github: ghClient,
		token:  token,
	}
}

// GitHub returns the underlying go-github client for direct API access.
func (c *Client) GitHub() *github.Client {
	return c.github
}

// GetAuthenticatedUser returns the currently authenticated GitHub user.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (*github.User, error) {
	user, _, err := c.github.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}
	return user, nil
}

// ValidateToken checks whether the configured token has valid GitHub API access
// by making a lightweight authenticated request. Returns a descriptive error
// if the token is expired, revoked, or lacks sufficient permissions.
//
// TODO: consider also checking rate limit headers here to surface quota info.
// TODO: surface the authenticated username in a debug log on success.
func (c *Client) ValidateToken(ctx context.Context) error {
	_, err := c.GetAuthenticatedUser(ctx)
	return err
}

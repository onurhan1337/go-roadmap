package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github-user-activity/models"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPIBaseURL     = "https://api.github.com"
	defaultTimeout       = 10 * time.Second
	maxRetries           = 3
	retryDelay           = time.Second
	headerRateLimit      = "X-RateLimit-Remaining"
	headerRateLimitReset = "X-RateLimit-Reset"
)

type RateLimit struct {
	Remaining int
	ResetAt   time.Time
}

type Client struct {
	httpClient *http.Client
	token      string
	maxRetries int
	retryDelay time.Duration
	rateLimit  *RateLimit
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		token:      token,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (c *Client) checkRateLimit(resp *http.Response) {
	remaining := resp.Header.Get(headerRateLimit)
	reset := resp.Header.Get(headerRateLimitReset)

	if remaining != "" {
		if rem, err := strconv.Atoi(remaining); err == nil {
			if c.rateLimit == nil {
				c.rateLimit = &RateLimit{}
			}
			c.rateLimit.Remaining = rem
		}
	}

	if reset != "" {
		if resetTime, err := strconv.ParseInt(reset, 10, 64); err == nil {
			if c.rateLimit == nil {
				c.rateLimit = &RateLimit{}
			}
			c.rateLimit.ResetAt = time.Unix(resetTime, 0)
		}
	}
}

func (c *Client) GetRateLimit() *RateLimit {
	return c.rateLimit
}

func (c *Client) retryRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		newReq := req.Clone(ctx)

		resp, err := c.httpClient.Do(newReq)
		if err == nil {
			if resp.StatusCode < 500 {
				return resp, nil
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt < c.maxRetries {
			backoff := time.Duration(attempt+1) * c.retryDelay

			timer := time.NewTimer(backoff)

			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
				fmt.Printf("Retry attempt %d/%d after %v delay\n",
					attempt+1, c.maxRetries, backoff)
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

func (c *Client) FetchUserEvents(ctx context.Context, username string) ([]models.GithubEvent, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if c.rateLimit != nil && c.rateLimit.Remaining == 0 {
		waitTime := time.Until(c.rateLimit.ResetAt)
		if waitTime > 0 {
			return nil, fmt.Errorf("rate limit exceeded, resets in %v", waitTime.Round(time.Second))
		}
	}

	url := fmt.Sprintf("%s/users/%s/events", githubAPIBaseURL, username)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	response, err := c.retryRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer response.Body.Close()

	c.checkRateLimit(response)

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("API request failed with status code: %d: %s",
			response.StatusCode, string(body))
	}

	var events []models.GithubEvent
	if err := json.NewDecoder(response.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return events, nil
}

func validateUsername(username string) error {
	// Remove leading and trailing whitespace from the username
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Validate username length according to GitHub's rules:
	// - Minimum length: 1 character
	// - Maximum length: 39 characters
	if len(username) < models.GithubRules.MinLength || len(username) > models.GithubRules.MaxLength {
		return fmt.Errorf("username length must be between %d and %d characters",
			models.GithubRules.MinLength, models.GithubRules.MaxLength)
	}

	// Validate username format according to GitHub's rules:
	// - Cannot start with a hyphen (-)
	// - Cannot end with a hyphen (-)
	// - Cannot contain consecutive hyphens (--)
	if strings.HasPrefix(username, "-") || strings.HasSuffix(username, "-") ||
		strings.Contains(username, "--") {
		return fmt.Errorf("invalid username format: cannot start/end with hyphen or contain consecutive hyphens")
	}

	return nil
}

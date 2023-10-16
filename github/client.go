package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"log/slog"

	"github.com/liweiyi88/trendshift-backend/model"
)

var ErrNotFound = errors.New("not found on GitHub.")
var ErrAccessBlocked = errors.New("repository access blocked.")

// GitHub rest api client
type Client struct {
	Token string // the personal acesss token, if set, the common rate limit is 5000 reqs/hour, otherwise, it will be 60 reqs/hour.
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
	}
}

func (ghClient *Client) GetDeveloper(ctx context.Context, username string) (model.Developer, error) {
	url := fmt.Sprintf("%s/%s", "https://api.github.com/users", username)

	var developer model.Developer

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return developer, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	if strings.TrimSpace(ghClient.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+ghClient.Token)
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return developer, fmt.Errorf("failed to send get developer request %v", err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			slog.Info("failed to close response body when fetch developer:", err)
		}
	}()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return developer, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &developer)

	if err != nil {
		return developer, fmt.Errorf("failed to decode developer body: %v", err)
	}

	slog.Info(fmt.Sprintf("fetching %s", developer.Username), slog.Group("github",
		slog.String("X-Ratelimit-Limit", res.Header.Get("X-Ratelimit-Limit")),
		slog.String("X-Ratelimit-Remaining", res.Header.Get("X-Ratelimit-Remaining")),
		slog.String("X-Ratelimit-Reset", res.Header.Get("X-Ratelimit-Reset")),
	))

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return developer, ErrNotFound
		}

		if res.StatusCode == http.StatusUnavailableForLegalReasons {
			return developer, ErrAccessBlocked
		}

		return developer, fmt.Errorf("request %s is not successful, get status code: %d, body: %s", url, res.StatusCode, string(body))
	}

	return developer, nil
}

func (ghClient *Client) GetRepository(ctx context.Context, fullName string) (model.GhRepository, error) {
	url := fmt.Sprintf("%s/%s", "https://api.github.com/repos", fullName)

	var ghRepository model.GhRepository

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ghRepository, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	if strings.TrimSpace(ghClient.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+ghClient.Token)
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return ghRepository, fmt.Errorf("failed to send get repository request %v", err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			slog.Info("failed to close response body when fetch repository:", err)
		}
	}()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return ghRepository, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &ghRepository)

	if err != nil {
		return ghRepository, fmt.Errorf("failed to decode repository body: %v", err)
	}

	slog.Info(fmt.Sprintf("fetching %s", ghRepository.FullName), slog.Group("github",
		slog.String("X-Ratelimit-Limit", res.Header.Get("X-Ratelimit-Limit")),
		slog.String("X-Ratelimit-Remaining", res.Header.Get("X-Ratelimit-Remaining")),
		slog.String("X-Ratelimit-Reset", res.Header.Get("X-Ratelimit-Reset")),
	))

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return ghRepository, ErrNotFound
		}

		if res.StatusCode == http.StatusUnavailableForLegalReasons {
			return ghRepository, ErrAccessBlocked
		}

		return ghRepository, fmt.Errorf("request %s is not successful, get status code: %d, body: %s", url, res.StatusCode, string(body))
	}

	return ghRepository, nil
}

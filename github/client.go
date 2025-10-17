package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/liweiyi88/trendshift-backend/model"
)

var ErrNotFound = errors.New("not found on GitHub")
var ErrAccessBlocked = errors.New("repository access blocked")
var ErrTooManyRequests = errors.New("too many requests")

const GraphQLURL = "https://api.github.com/graphql"

type Stargazer struct {
	StarredAt time.Time
	Login     string
}

type GraphQLResponse struct {
	Data struct {
		Repository struct {
			Stargazers struct {
				Edges []struct {
					StarredAt string `json:"starredAt"`
					Node      struct {
						Login string `json:"login"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"stargazers"`
		} `json:"repository"`
	} `json:"data"`
}

func printRateLimitHeaders(name string, res http.Response) {
	slog.Info(name, slog.Group("github",
		slog.String("X-Ratelimit-Limit", res.Header.Get("X-Ratelimit-Limit")),
		slog.String("X-Ratelimit-Remaining", res.Header.Get("X-Ratelimit-Remaining")),
		slog.String("X-Ratelimit-Reset", res.Header.Get("X-Ratelimit-Reset")),
	))
}

func checkGitHubResponse(res *http.Response, body []byte, context string) error {
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnavailableForLegalReasons, http.StatusForbidden:
		return ErrAccessBlocked
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	default:
		return fmt.Errorf("[%s] request failed: status=%d, body=%s", context, res.StatusCode, string(body))
	}
}

type Client struct {
	Token string // the personal acesss token, if set, the common rate limit is 5000 reqs/hour, otherwise, it will be 60 reqs/hour.
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
	}
}

func (ghClient *Client) GetRepositoryStars(
	ctx context.Context,
	owner, repo string,
	cursor *string,
	start *time.Time,
	end *time.Time) ([]Stargazer, *string, error) {
	query := `
query ($owner: String!, $repo: String!, $after: String) {
  repository(owner: $owner, name: $repo) {
    stargazers(first: 100, after: $after, orderBy: {field: STARRED_AT, direction: DESC}) {
      edges {
        starredAt
        node {
          login
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`

	variables := map[string]any{
		"owner": owner,
		"repo":  repo,
		"after": cursor,
	}

	requestData := map[string]any{
		"query":     query,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(requestData)
	if err != nil {
		return nil, nil, fmt.Errorf("[stargazers] failed to marshal request data, %v", requestData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GraphQLURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("[stargazers] failed to create new http request, error: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ghClient.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, nil, fmt.Errorf("[stargazers] failed to send graphql request, error: %v", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("[stargazers] failed to close response body", slog.String("error", err.Error()))
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("[stargazers] failed to read response body, error: %v", err)
	}

	err = checkGitHubResponse(res, body, "stargazers")
	if err != nil {
		return nil, nil, err
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, nil, fmt.Errorf("[stargazers] failed to unmarshal graphql response, error: %v", err)
	}

	printRateLimitHeaders("get stargazers", *res)

	stars := make([]Stargazer, 0, len(gqlResp.Data.Repository.Stargazers.Edges))

	for _, edge := range gqlResp.Data.Repository.Stargazers.Edges {
		starredAt, err := time.Parse(time.RFC3339, edge.StarredAt)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse starredAt %s, error: %v", edge.StarredAt, err)
		}

		if start != nil && starredAt.Before(*start) {
			return stars, nil, nil
		}

		if end != nil && starredAt.After(*end) {
			return stars, nil, nil
		}

		stargazer := Stargazer{
			Login:     edge.Node.Login,
			StarredAt: starredAt,
		}

		stars = append(stars, stargazer)
	}

	var nextCursor *string
	if gqlResp.Data.Repository.Stargazers.PageInfo.HasNextPage {
		nextCursor = &gqlResp.Data.Repository.Stargazers.PageInfo.EndCursor
	}

	return stars, nextCursor, nil
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
			slog.Any("failed to close response body when fetch developer:", err)
		}
	}()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return developer, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &developer)

	if err != nil {
		return developer, fmt.Errorf("failed to decode developer body err: %v, received: %s, status code: %s", err, string(body), res.Status)
	}

	printRateLimitHeaders(developer.Username, *res)

	return developer, checkGitHubResponse(res, body, "developer")
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
			slog.Any("failed to close response body when fetch repository:", err)
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

	printRateLimitHeaders(ghRepository.FullName, *res)

	return ghRepository, checkGitHubResponse(res, body, "repository")
}

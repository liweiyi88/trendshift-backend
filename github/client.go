package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

type Fork struct {
	CreatedAt time.Time
	Login     string
}

type Pr struct {
	Number   int
	Title    string
	MergedAt time.Time
}

type Issue struct {
	Title     string
	Number    int
	Closed    bool
	ClosedAt  time.Time
	CreatedAt time.Time
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
			Issues struct {
				Edges []struct {
					Node struct {
						Number    int    `json:"number"`
						Title     string `json:"title"`
						Closed    bool   `json:"closed"`
						ClosedAt  string `json:"closedAt"`
						CreatedAt string `json:"createdAt"`
						UpdatedAt string `json:"updatedAt"`
						Author    struct {
							Login string `json:"login"`
						} `json:"author"`
						URL string `json:"url"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"issues"`
			PullRequests struct {
				Edges []struct {
					Node struct {
						Number   int    `json:"number"`
						Title    string `json:"title"`
						MergedAt string `json:"mergedAt"`
						Author   struct {
							Login string `json:"login"`
						} `json:"author"`
						URL string `json:"url"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"pullRequests"`
			Forks struct {
				Edges []struct {
					Node struct {
						Name  string `json:"name"`
						Owner struct {
							Login string `json:"login"`
						} `json:"owner"`
						CreatedAt string `json:"createdAt"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"forks"`
		} `json:"repository"`
	} `json:"data"`
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
	TokenPool *TokenPool
}

func NewClient(tokenPool *TokenPool) *Client {
	return &Client{
		TokenPool: tokenPool,
	}
}

func fetch[T any](
	ctx context.Context,
	query string,
	tokenPool *TokenPool,
	owner, repo string,
	cursor *string,
	extractEdges func(body []byte) ([]T, *string, error),
) ([]T, *string, error) {
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
		return nil, nil, fmt.Errorf("[github graphql] failed to marshal request data, %v", requestData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", GraphQLURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("[github graphql] failed to create new http request, error: %v", err)
	}

	token, err := tokenPool.GetToken()
	if err != nil {
		return nil, nil, fmt.Errorf("[github graphql] failed to get token from token pool, error: %w", err)
	}

	// Also allow to send request without token
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, nil, fmt.Errorf("[github graphql] failed to send graphql request, error: %v", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("failed to close response body", slog.String("error", err.Error()))
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body, error: %v", err)
	}

	err = checkGitHubResponse(res, body, "github graphql")
	if err != nil {
		return nil, nil, err
	}

	slog.Debug("fetching repo monthly date", slog.String("repository", fmt.Sprintf("%s/%s", owner, repo)))

	remainingStr := res.Header.Get("X-Ratelimit-Remaining")
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)

	if err != nil {
		return nil, nil, fmt.Errorf("[github graphql] failed to parse X-Ratelimit-Remaining to int")
	}

	resetAt := res.Header.Get("X-Ratelimit-Reset")
	tokenPool.Update(token, int(remaining), resetAt)

	return extractEdges(body)
}

func (ghClient *Client) GetRepositoryForks(
	ctx context.Context,
	owner, repo string,
	cursor *string,
	start *time.Time,
	end *time.Time) ([]Fork, *string, error) {
	query := `
query ($owner: String!, $repo: String!, $after: String) {
  repository(owner: $owner, name: $repo) {
    forks(first: 100, after: $after, orderBy: {field: CREATED_AT, direction: DESC}) {
      edges {
        node {
		  name
		  createdAt
		  owner {
		    login
		  }
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`

	extractEdges := func(body []byte) ([]Fork, *string, error) {
		var gqlResp GraphQLResponse
		if err := json.Unmarshal(body, &gqlResp); err != nil {
			return nil, nil, fmt.Errorf("[forks] failed to unmarshal graphql response, error: %v", err)
		}

		forks := make([]Fork, 0, len(gqlResp.Data.Repository.Forks.Edges))

		for _, edge := range gqlResp.Data.Repository.Forks.Edges {
			createdAt, err := time.Parse(time.RFC3339, edge.Node.CreatedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse createdAt %s, error: %v", edge.Node.CreatedAt, err)
			}

			if start != nil && createdAt.Before(*start) {
				return forks, nil, nil
			}

			if end != nil && createdAt.After(*end) {
				return forks, nil, nil
			}

			fork := Fork{
				Login:     edge.Node.Owner.Login,
				CreatedAt: createdAt,
			}

			forks = append(forks, fork)
		}

		var nextCursor *string
		if gqlResp.Data.Repository.Forks.PageInfo.HasNextPage {
			nextCursor = &gqlResp.Data.Repository.Forks.PageInfo.EndCursor
		}

		return forks, nextCursor, nil
	}

	return fetch(ctx, query, ghClient.TokenPool, owner, repo, cursor, extractEdges)
}

func (ghClient *Client) GetIssues(
	ctx context.Context,
	owner, repo string,
	cursor *string,
	start, end *time.Time) ([]Issue, *string, error) {
	query := `
query ($owner: String!, $repo: String!, $after: String) {
  repository(owner: $owner, name: $repo) {
    issues(first: 100, after: $after, orderBy: {field: UPDATED_AT, direction: DESC}) {
      edges {
	    node {
		  number
          title
		  closed
		  createdAt
		  updatedAt
          closedAt
          author {
            login
          }
          url
		}
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`

	extractEdges := func(body []byte) ([]Issue, *string, error) {
		var gqlResp GraphQLResponse
		if err := json.Unmarshal(body, &gqlResp); err != nil {
			return nil, nil, fmt.Errorf("[issues] failed to unmarshal graphql response, error: %v", err)
		}

		issues := make([]Issue, 0, len(gqlResp.Data.Repository.Issues.Edges))

		for _, edge := range gqlResp.Data.Repository.Issues.Edges {
			createdAt, err := time.Parse(time.RFC3339, edge.Node.CreatedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[issues] failed to parse createdAt %s, error: %v", createdAt, err)
			}

			updatedAt, err := time.Parse(time.RFC3339, edge.Node.UpdatedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[issues] failed to parse updatedAt %s, error: %v", updatedAt, err)
			}

			var closedAt time.Time

			if edge.Node.ClosedAt != "" {
				t, err := time.Parse(time.RFC3339, edge.Node.ClosedAt)
				if err != nil {
					return nil, nil, fmt.Errorf("[issues] failed to parse closedAt %s, error: %v", closedAt, err)
				}

				closedAt = t
			}

			if start != nil && updatedAt.Before(*start) {
				return issues, nil, nil
			}

			if end != nil && updatedAt.After(*end) {
				return issues, nil, nil
			}

			issue := Issue{
				Title:     edge.Node.Title,
				Number:    edge.Node.Number,
				Closed:    edge.Node.Closed,
				CreatedAt: createdAt,
				ClosedAt:  closedAt,
			}

			issues = append(issues, issue)
		}

		var nextCursor *string
		if gqlResp.Data.Repository.Issues.PageInfo.HasNextPage {
			nextCursor = &gqlResp.Data.Repository.Issues.PageInfo.EndCursor
		}

		return issues, nextCursor, nil
	}

	return fetch(ctx, query, ghClient.TokenPool, owner, repo, cursor, extractEdges)
}

func (ghClient *Client) GetMergedPrs(
	ctx context.Context,
	owner,
	repo string,
	cursor *string, start, end *time.Time) ([]Pr, *string, error) {
	query := `
query ($owner: String!, $repo: String!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequests(first: 100, after: $after, states: MERGED, orderBy: {field: UPDATED_AT, direction: DESC}) {
      edges {
	    node {
		  number
          title
          mergedAt
          author {
            login
          }
          url
		}
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}`
	extractEdges := func(body []byte) ([]Pr, *string, error) {
		var gqlResp GraphQLResponse
		if err := json.Unmarshal(body, &gqlResp); err != nil {
			return nil, nil, fmt.Errorf("[prs] failed to unmarshal graphql response, error: %v", err)
		}

		prs := make([]Pr, 0, len(gqlResp.Data.Repository.PullRequests.Edges))

		for _, edge := range gqlResp.Data.Repository.PullRequests.Edges {
			mergedAt, err := time.Parse(time.RFC3339, edge.Node.MergedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[prs] failed to parse starredAt %s, error: %v", mergedAt, err)
			}

			if start != nil && mergedAt.Before(*start) {
				return prs, nil, nil
			}

			if end != nil && mergedAt.After(*end) {
				return prs, nil, nil
			}

			pr := Pr{
				Title:    edge.Node.Title,
				Number:   edge.Node.Number,
				MergedAt: mergedAt,
			}

			prs = append(prs, pr)
		}

		var nextCursor *string
		if gqlResp.Data.Repository.PullRequests.PageInfo.HasNextPage {
			nextCursor = &gqlResp.Data.Repository.PullRequests.PageInfo.EndCursor
		}

		return prs, nextCursor, nil
	}

	return fetch(ctx, query, ghClient.TokenPool, owner, repo, cursor, extractEdges)
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

	extractEdges := func(body []byte) ([]Stargazer, *string, error) {
		var gqlResp GraphQLResponse
		if err := json.Unmarshal(body, &gqlResp); err != nil {
			return nil, nil, fmt.Errorf("[stargazers] failed to unmarshal graphql response, error: %v", err)
		}

		stars := make([]Stargazer, 0, len(gqlResp.Data.Repository.Stargazers.Edges))

		for _, edge := range gqlResp.Data.Repository.Stargazers.Edges {
			starredAt, err := time.Parse(time.RFC3339, edge.StarredAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[stargazers] failed to parse starredAt %s, error: %v", edge.StarredAt, err)
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

	return fetch(ctx, query, ghClient.TokenPool, owner, repo, cursor, extractEdges)
}

func (ghClient *Client) GetDeveloper(ctx context.Context, username string) (model.Developer, error) {
	url := fmt.Sprintf("%s/%s", "https://api.github.com/users", username)

	var developer model.Developer

	token, err := ghClient.TokenPool.GetToken()
	if err != nil {
		return developer, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return developer, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Also allow to send request without token
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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

	remainingStr := res.Header.Get("X-Ratelimit-Remaining")
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)

	if err != nil {
		return developer, fmt.Errorf("[github graphql] failed to parse X-Ratelimit-Remaining to int")
	}

	resetAt := res.Header.Get("X-Ratelimit-Reset")
	ghClient.TokenPool.Update(token, int(remaining), resetAt)

	return developer, checkGitHubResponse(res, body, "developer")
}

func (ghClient *Client) GetRepository(ctx context.Context, fullName string) (model.GhRepository, error) {
	url := fmt.Sprintf("%s/%s", "https://api.github.com/repos", fullName)

	var ghRepository model.GhRepository

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ghRepository, err
	}

	token, err := ghClient.TokenPool.GetToken()
	if err != nil {
		return ghRepository, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Also allow to send reqeusts without token
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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

	remainingStr := res.Header.Get("X-Ratelimit-Remaining")
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)

	if err != nil {
		return ghRepository, fmt.Errorf("[github graphql] failed to parse X-Ratelimit-Remaining to int")
	}

	resetAt := res.Header.Get("X-Ratelimit-Reset")
	ghClient.TokenPool.Update(token, int(remaining), resetAt)

	return ghRepository, checkGitHubResponse(res, body, "repository")
}

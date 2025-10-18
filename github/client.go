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

func fetchPaginated[T any](
	ctx context.Context,
	query string,
	token string,
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

	req.Header.Set("Authorization", "Bearer "+token)
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

	printRateLimitHeaders("github graphql", *res)
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

	return fetchPaginated(ctx, query, ghClient.Token, owner, repo, cursor, extractEdges)
}

func (ghClient *Client) GetClosedIssues(
	ctx context.Context,
	owner, repo string,
	cursor *string,
	start, end *time.Time) ([]Issue, *string, error) {
	query := `
query ($owner: String!, $repo: String!, $after: String) {
  repository(owner: $owner, name: $repo) {
    issues(first: 100, after: $after, states: CLOSED, orderBy: {field: UPDATED_AT, direction: DESC}) {
      edges {
	    node {
		  number
          title
		  closed
		  createdAt
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
			return nil, nil, fmt.Errorf("[closed issue] failed to unmarshal graphql response, error: %v", err)
		}

		closedIssues := make([]Issue, 0, len(gqlResp.Data.Repository.Issues.Edges))

		for _, edge := range gqlResp.Data.Repository.Issues.Edges {

			createdAt, err := time.Parse(time.RFC3339, edge.Node.CreatedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[closed issue] failed to parse createdAt %s, error: %v", createdAt, err)
			}

			closedAt, err := time.Parse(time.RFC3339, edge.Node.ClosedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("[closed issue] failed to parse closedAt %s, error: %v", closedAt, err)
			}

			if start != nil && closedAt.Before(*start) {
				return closedIssues, nil, nil
			}

			if end != nil && closedAt.After(*end) {
				return closedIssues, nil, nil
			}

			issue := Issue{
				Title:     edge.Node.Title,
				Number:    edge.Node.Number,
				Closed:    edge.Node.Closed,
				CreatedAt: createdAt,
				ClosedAt:  closedAt,
			}

			closedIssues = append(closedIssues, issue)
		}

		var nextCursor *string
		if gqlResp.Data.Repository.Issues.PageInfo.HasNextPage {
			nextCursor = &gqlResp.Data.Repository.Issues.PageInfo.EndCursor
		}

		return closedIssues, nextCursor, nil
	}

	return fetchPaginated(ctx, query, ghClient.Token, owner, repo, cursor, extractEdges)
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

	return fetchPaginated(ctx, query, ghClient.Token, owner, repo, cursor, extractEdges)
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

	return fetchPaginated(ctx, query, ghClient.Token, owner, repo, cursor, extractEdges)
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

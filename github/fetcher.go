package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/exp/slog"
)

// Fetch github repo details via REST api and store the result in DB.
type Fetcher struct {
	Pat              string // the personal acesss token, if set, the common rate limit is 5000 req/hour, otherwise, it will be 60 req/hour.
	GhRepositoryRepo GhRepositoryRepo
}

func (f *Fetcher) FetchRepository(ctx context.Context, fullName string) (GhRepository, error) {
	url := fmt.Sprintf("%s/%s", "https://api.github.com/repos", fullName)

	var ghRepository GhRepository

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ghRepository, err
	}

	if strings.TrimSpace(f.Pat) != "" {
		req.Header.Set("Authorization: Bearer", f.Pat)
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

	slog.Info("Fetch repo response", slog.Group("github",
		slog.String("X-Ratelimit-Limit", res.Header.Get("X-Ratelimit-Limit")),
		slog.String("X-Ratelimit-Remaining", res.Header.Get("X-Ratelimit-Remaining")),
		slog.String("X-Ratelimit-Reset", res.Header.Get("X-Ratelimit-Reset")),
	))

	if res.StatusCode != http.StatusOK {
		return ghRepository, fmt.Errorf("request %s is not successful, get status code: %d, body: %s", url, res.StatusCode, string(body))
	}

	return ghRepository, nil
}

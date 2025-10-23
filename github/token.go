package github

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrTokenNotAvailable = errors.New("all github tokens are not available")

type token struct {
	token     string
	remaining int
	resetAt   time.Time
}

func (t *token) isAvailable() bool {
	return t.remaining > 0 || time.Now().After(t.resetAt)
}

type TokenPool struct {
	mu              sync.Mutex
	tokens          []*token
	allowEmptyToken bool
	index           int
}

type Option func(tp *TokenPool)

func WithAllowEmptytoken(allow bool) Option {
	return func(tp *TokenPool) {
		tp.allowEmptyToken = allow
	}
}

func NewTokenPool(tokens []string, opts ...Option) *TokenPool {
	tp := &TokenPool{}

	for _, opt := range opts {
		opt(tp)
	}

	for _, t := range tokens {
		tp.tokens = append(tp.tokens, &token{
			token:     t,
			remaining: 5000,
			resetAt:   time.Now(),
		})
	}

	return tp
}

func (tp *TokenPool) Update(token string, remaining int, resetAt time.Time) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for _, t := range tp.tokens {
		if t.token == token {
			t.remaining = remaining
			t.resetAt = resetAt
			return
		}
	}
}

func (tp *TokenPool) GetToken() (string, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for i, token := range tp.tokens {
		if token.isAvailable() {
			tp.index = i
			return token.token, nil
		}
	}

	if tp.allowEmptyToken {
		return "", nil
	}

	return "", ErrTokenNotAvailable
}

func (tp *TokenPool) EarliestReset() time.Time {
	var earliest time.Time
	for _, t := range tp.tokens {
		if earliest.IsZero() || t.resetAt.Before(earliest) {
			earliest = t.resetAt
		}
	}

	return earliest
}

func (tp *TokenPool) Debug() {
	group := slog.Group("github_tokens")

	for index, token := range tp.tokens {
		slog.Debug("github tokens availability",
			slog.Int("index", index),
			slog.Int("remaining", token.remaining),
			slog.String("resetAt", token.resetAt.Format(time.DateTime)),
			group,
		)
	}
}

package helpers

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

type GitHubRoundTripper struct {
	OauthClient *http.Client
}

func NewGitHubRoundTripper(ctx context.Context, token string) (*GitHubRoundTripper, error) {
	inst := new(GitHubRoundTripper)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	inst.OauthClient = oauth2.NewClient(ctx, ts)
	inst.OauthClient.Transport = inst
	return inst, nil
}

func (g *GitHubRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := g.OauthClient.Do(r)
	limit := resp.Header.Get("X-RateLimit-Limit")
	used := resp.Header.Get("X-RateLimit-Used")
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")
	// convert from timestamp int string format (reset) to time.Duration
	resetTime, err := time.Parse("2006-01-02T15:04:05-0700", reset)
	if err != nil {
		return resp, fmt.Errorf("unable to parse 'X-RateLimit-Reset' header. error: %v", err)
	}

	log.Printf("GitHub API usage limit statistics: [%v/%v]...", used, limit)

	if err != nil || resp.StatusCode == http.StatusForbidden {
		if remaining == "0" {
			log.Printf("GitHub API usage limits exceeded... Waiting for reset: %s...", resetTime.String())
			<-time.After(resetTime.Sub(time.Now()))
			return g.RoundTrip(r)
		}
	}
	return resp, err
}

package api

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pcanilho/gh-tidy/models"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"os"
	"strings"
	"time"
)

type GitHub struct {
	clientV3 *github.Client
	clientV4 *githubv4.Client
}

func NewSession() (*GitHub, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if len(strings.TrimSpace(token)) == 0 {
		return nil, fmt.Errorf("a GITHUB_TOKEN environment variable needs to be set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &GitHub{clientV3: github.NewClient(tc), clientV4: githubv4.NewClient(tc)}, nil
}
func (gh *GitHub) ListPRs(ctx context.Context, states []string, owner, repo string) ([]*models.GitHubPR, error) {
	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []struct {
					Id      string
					Commits struct {
						Nodes []struct {
							Commit struct {
								CommittedDate time.Time
							}
							PullRequest struct {
								Number int
								Url    string
							}
						}
					} `graphql:"commits(last: 1)"`
					BaseRefName string
					HeadRefName string
				}
			} `graphql:"pullRequests(states: $states, last: $last)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	var sts []githubv4.PullRequestState
	for _, s := range states {
		sts = append(sts, githubv4.PullRequestState(strings.ToUpper(s)))
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(repo),
		"last":   githubv4.Int(100),
		"states": sts,
	}

	if err := gh.clientV4.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	var out []*models.GitHubPR
	for _, pr := range query.Repository.PullRequests.Nodes {
		out = append(out, &models.GitHubPR{
			Source:         pr.HeadRefName,
			Target:         pr.BaseRefName,
			LastCommitDate: pr.Commits.Nodes[0].Commit.CommittedDate,
			Id:             pr.Id,
			Number:         pr.Commits.Nodes[0].PullRequest.Number,
			Url:            pr.Commits.Nodes[0].PullRequest.Url,
		})
	}
	return out, nil
}

func (gh *GitHub) ListBranches(ctx context.Context, owner, repo string) ([]*models.GitHubBranch, error) {
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Id     string
					Name   string
					Target struct {
						Commit struct {
							CommittedDate time.Time
						} `graphql:"... on Commit"`
					}
				}
			} `graphql:"refs(last: $last, refPrefix: $refPrefix)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"name":      githubv4.String(repo),
		"last":      githubv4.Int(100),
		"refPrefix": githubv4.String("refs/heads/"),
	}

	if err := gh.clientV4.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	var out []*models.GitHubBranch
	for _, n := range query.Repository.Refs.Nodes {
		out = append(out, &models.GitHubBranch{Name: n.Name, LastCommitDate: n.Target.Commit.CommittedDate, Id: n.Id})
	}
	return out, nil
}

func (gh *GitHub) DeleteRefs(ctx context.Context, refs ...string) ([]string, error) {
	if refs == nil || len(refs) == 0 {
		return nil, fmt.Errorf("no refs have been specified")
	}

	var mutation struct {
		DeleteRef struct {
			Typename string `graphql:"typename :__typename"`
		} `graphql:"deleteRef(input: {refId: $input})"`
	}

	var out []string
	for _, ref := range refs {
		if err := gh.clientV4.Mutate(ctx, &mutation, githubv4.Input(ref), nil); err != nil {
			return out, err
		}
		out = append(out, mutation.DeleteRef.Typename)
	}

	return out, nil
}

func (gh *GitHub) ClosePRs(ctx context.Context, ids ...string) ([]string, error) {
	if ids == nil || len(ids) == 0 {
		return nil, fmt.Errorf("no PR ids have been specified")
	}

	var mutation struct {
		ClosePullRequest struct {
			Typename string `graphql:"typename :__typename"`
		} `graphql:"closePullRequest(input: {pullRequestId: $input})"`
	}

	var out []string
	for _, id := range ids {
		if err := gh.clientV4.Mutate(ctx, &mutation, githubv4.Input(id), nil); err != nil {
			return out, err
		}
		out = append(out, mutation.ClosePullRequest.Typename)
	}

	return out, nil
}
package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pcanilho/gh-tidy/api/helpers"
	"github.com/shurcooL/githubv4"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const _defaultWorkerCount = 20

type GitHub struct {
	enterpriseEndpoint string
	clientV3           *github.Client
	clientV4           *githubv4.Client

	httpClient  *http.Client
	context     context.Context
	workerCount int
}

type Option = func(*GitHub)

func WithHttpClient(httpClient *http.Client) Option {
	return func(session *GitHub) {
		session.httpClient = httpClient
	}
}

func WithContext(ctx context.Context) Option {
	return func(session *GitHub) {
		session.context = ctx
	}
}

func WithEnterpriseEndpoint(enterpriseEndpoint string) Option {
	return func(session *GitHub) {
		session.enterpriseEndpoint = enterpriseEndpoint
	}
}

func WithWorkerCount(workerCount int) Option {
	return func(session *GitHub) {
		session.workerCount = workerCount
	}
}

func NewSession(opts ...Option) (*GitHub, error) {
	inst := new(GitHub)
	for _, opt := range opts {
		opt(inst)
	}

	if inst.context == nil {
		inst.context = context.Background()
	}

	if inst.workerCount == 0 {
		inst.workerCount = _defaultWorkerCount
	}

	token := os.Getenv("GITHUB_TOKEN")
	if len(strings.TrimSpace(token)) == 0 {
		return nil, fmt.Errorf("a GITHUB_TOKEN environment variable needs to be set")
	}

	if inst.httpClient == nil {
		ghRoundTripper, err := helpers.NewGitHubRoundTripper(inst.context, token)
		if err != nil {
			return nil, err
		}
		inst.httpClient = ghRoundTripper.OauthClient
	}
	if len(inst.enterpriseEndpoint) != 0 {
		clientV3, err := github.NewEnterpriseClient(inst.enterpriseEndpoint, inst.enterpriseEndpoint, inst.httpClient)
		if err != nil {
			return nil, err
		}
		inst.clientV3 = clientV3
		inst.clientV4 = githubv4.NewEnterpriseClient(inst.enterpriseEndpoint, inst.httpClient)
	} else {
		inst.clientV3 = github.NewClient(inst.httpClient)
		inst.clientV4 = githubv4.NewClient(inst.httpClient)
	}
	return inst, nil
}
func (gh *GitHub) ListPRs(ctx context.Context, states []string, owner, repo string) ([]*GitHubPR, error) {
	if len(owner) == 0 {
		return nil, fmt.Errorf("an owner must be specified")
	}

	if len(repo) == 0 {
		return nil, fmt.Errorf("a repo must be specified")
	}

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
				PageInfo struct {
					EndCursor   string
					HasNextPage bool
				}
			} `graphql:"pullRequests(first: $first, after: $after, states: $states)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	var sts []githubv4.PullRequestState
	for _, s := range states {
		sts = append(sts, githubv4.PullRequestState(strings.ToUpper(s)))
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(repo),
		"first":  githubv4.Int(100),
		"after":  (*githubv4.String)(nil),
		"states": sts,
	}
	var out []*GitHubPR

	for {
		if err := gh.clientV4.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, pr := range query.Repository.PullRequests.Nodes {
			out = append(out, &GitHubPR{
				Source:         pr.HeadRefName,
				Target:         pr.BaseRefName,
				LastCommitDate: pr.Commits.Nodes[0].Commit.CommittedDate,
				Id:             pr.Id,
				Number:         pr.Commits.Nodes[0].PullRequest.Number,
				Url:            pr.Commits.Nodes[0].PullRequest.Url,
			})
		}
		if !query.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.String(query.Repository.PullRequests.PageInfo.EndCursor)
	}
	return out, nil
}

type RefType = string

const (
	BranchRefType RefType = "refs/heads/"
	TagRefType            = "refs/tags/"
)

func (gh *GitHub) ListRefs(ctx context.Context, owner, repo string, refType RefType) ([]*GitHubRef, error) {
	if len(owner) == 0 {
		return nil, fmt.Errorf("an owner must be specified")
	}

	if len(repo) == 0 {
		return nil, fmt.Errorf("a repo must be specified")
	}

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
						Tag struct {
							Tagger struct {
								Date time.Time
							}
						} `graphql:"... on Tag"`
					}
				}
				PageInfo struct {
					EndCursor   string
					HasNextPage bool
				}
			} `graphql:"refs(first: $first, after: $after, refPrefix: $refPrefix)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"name":      githubv4.String(repo),
		"refPrefix": githubv4.String(refType),
		"first":     githubv4.Int(100),
		"after":     (*githubv4.String)(nil),
	}

	var out []*GitHubRef
	for {
		if err := gh.clientV4.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, n := range query.Repository.Refs.Nodes {
			commitDate := n.Target.Commit.CommittedDate
			tagDate := n.Target.Tag.Tagger.Date

			model := &GitHubRef{Name: n.Name, Id: n.Id}
			if !commitDate.IsZero() {
				model.LastCommitDate = &commitDate
			}
			if !tagDate.IsZero() {
				model.TagDate = &tagDate
			}
			out = append(out, model)
		}

		if !query.Repository.Refs.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.String(query.Repository.Refs.PageInfo.EndCursor)
	}
	return out, nil
}

func (gh *GitHub) DeleteRefs(ctx context.Context, refs ...string) error {
	if refs == nil || len(refs) == 0 {
		return fmt.Errorf("no refs have been specified")
	}

	var mutation struct {
		DeleteRef struct {
			Typename string `graphql:"typename :__typename"`
		} `graphql:"deleteRef(input: {refId: $input})"`
	}

	ec := make(chan error, len(refs))
	sem := make(chan struct{}, gh.workerCount)

	var wg sync.WaitGroup
	wg.Add(len(refs))
	go func() {
		wg.Wait()
		close(ec)
		close(sem)
	}()

	for _, ref := range refs {
		sem <- struct{}{}
		go func(r string) {
			reqErr := gh.clientV4.Mutate(ctx, &mutation, githubv4.Input(r), nil)
			if reqErr != nil {
				ec <- fmt.Errorf("unable to delete ref: %v. error: %v", r, reqErr)
			}
			wg.Done()
			<-sem
		}(ref)
	}

	var err error
	for e := range ec {
		err = errors.Join(err, e)
	}
	return err
}

func (gh *GitHub) ClosePRs(ctx context.Context, ids ...string) error {
	if ids == nil || len(ids) == 0 {
		return fmt.Errorf("no PR ids have been specified")
	}

	var mutation struct {
		ClosePullRequest struct {
			Typename string `graphql:"typename :__typename"`
		} `graphql:"closePullRequest(input: {pullRequestId: $input})"`
	}

	ec := make(chan error, len(ids))
	sem := make(chan struct{}, gh.workerCount)

	var wg sync.WaitGroup
	wg.Add(len(ids))
	go func() {
		wg.Wait()
		close(ec)
		close(sem)
	}()

	for _, id := range ids {
		sem <- struct{}{}
		go func(identifier string) {
			reqErr := gh.clientV4.Mutate(ctx, &mutation, githubv4.Input(identifier), nil)
			if reqErr != nil {
				ec <- fmt.Errorf("unable to close PR: %v. error: %v", identifier, reqErr)
			}
			wg.Done()
			<-sem
		}(id)
	}

	var err error
	for e := range ec {
		err = errors.Join(err, e)
	}

	return err
}

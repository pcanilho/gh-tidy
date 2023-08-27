package api_test

import (
	"context"
	"fmt"
	"github.com/pcanilho/gh-tidy/api"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	envKey := "GITHUB_TOKEN"
	t.Run("session-w/o-token", func(ti *testing.T) {
		old := os.Getenv(envKey)
		assert.NoError(ti, os.Unsetenv(envKey))
		s, err := api.NewSession()

		assert.Error(ti, err)
		assert.Nil(ti, s)
		assert.NoError(ti, os.Setenv(envKey, old))
	})
	t.Run("session-w-token", func(ti *testing.T) {
		old := os.Getenv(envKey)
		assert.NoError(ti, os.Setenv(envKey, "XXX"))
		{
			s, err := api.NewSession()

			assert.NoError(ti, err)
			assert.NotNil(ti, s)
		}
		assert.NoError(ti, os.Unsetenv(envKey))
		if len(old) != 0 {
			assert.NoError(ti, os.Setenv(envKey, old))
		}
	})
}

func TestGitHub_ListRefs(t *testing.T) {
	envKey := "GITHUB_TOKEN"
	old := os.Getenv(envKey)
	assert.NoError(t, os.Setenv(envKey, "XXX"))

	setup(t)
	refType := new(api.RefType)
	t0 := "2023-08-29T19:20:49+01:00"
	t0p, terr := time.Parse(time.RFC3339, t0)
	owner, repo := "x", "y"
	assert.NoError(t, terr)
	assert.NotZero(t, t0p)

	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t,
			readBody(t, r),
			fmt.Sprintf(`{"query":"query($after:String$first:Int!$name:String!$owner:String!$refPrefix:String!){repository(owner: $owner, name: $name){refs(first: $first, after: $after, refPrefix: $refPrefix){nodes{id,name,target{... on Commit{committedDate},... on Tag{tagger{date}}}},pageInfo{endCursor,hasNextPage}}}}","variables":{"after":null,"first":100,"name":"%v","owner":"%v","refPrefix":"%v"}}`, repo, owner, *refType))
		writeBody(t, w, fmt.Sprintf(`{"data": {"repository": {"refs": {"nodes": [{"id": "007", "name": "test-ref", "target": {"committedDate": "%v", "tagger": {"date": "%v"}}}]}}}}`, t0, t0))
	})
	{
		*refType = api.BranchRefType
		t.Run("list-refs-branches-valid-match", func(ti *testing.T) {
			brs, err := ghApi.ListRefs(context.Background(), owner, repo, *refType)
			assert.NoError(ti, err)
			assert.Len(ti, brs, 1)
			expected := &api.GitHubRef{
				Id:             "007",
				Name:           "test-ref",
				LastCommitDate: &t0p,
				TagDate:        &t0p,
			}
			assert.Equal(ti, expected, brs[0])
		})
		t.Run("list-refs-branches-valid-mismatch", func(ti *testing.T) {
			brs, err := ghApi.ListRefs(context.Background(), owner, repo, *refType)
			assert.NoError(ti, err)
			assert.Len(ti, brs, 1)
			expected := &api.GitHubRef{
				Id:             "006",
				Name:           "test-ref",
				LastCommitDate: &t0p,
				TagDate:        &t0p,
			}
			assert.NotEqual(ti, expected, brs[0])
		})
		*refType = api.TagRefType
		t.Run("list-refs-tags-valid-match", func(ti *testing.T) {
			brs, err := ghApi.ListRefs(context.Background(), owner, repo, *refType)
			assert.NoError(ti, err)
			assert.Len(ti, brs, 1)
			expected := &api.GitHubRef{
				Id:             "007",
				Name:           "test-ref",
				LastCommitDate: &t0p,
				TagDate:        &t0p,
			}
			assert.Equal(ti, expected, brs[0])
		})
		t.Run("list-refs-invalid-owner", func(ti *testing.T) {
			owner = ""
			brs, err := ghApi.ListRefs(context.Background(), owner, repo, *refType)
			assert.Error(ti, err)
			assert.Nil(ti, brs)
		})
		t.Run("list-refs-invalid-repo", func(ti *testing.T) {
			repo = ""
			owner = "x"
			brs, err := ghApi.ListRefs(context.Background(), owner, repo, *refType)
			assert.Error(ti, err)
			assert.Nil(ti, brs)
		})
	}
	assert.NoError(t, os.Setenv(envKey, old))
}

func TestGitHub_ListPRs(t *testing.T) {
	envKey := "GITHUB_TOKEN"
	old := os.Getenv(envKey)
	assert.NoError(t, os.Setenv(envKey, "XXX"))

	setup(t)
	t0 := "2023-08-29T19:20:49+01:00"
	t0p, terr := time.Parse(time.RFC3339, t0)
	owner, repo := "x", "y"
	headName, baseName, url := "test-head-name", "test-base-name", "test-url"
	assert.NoError(t, terr)
	assert.NotZero(t, t0p)

	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t,
			readBody(t, r),
			fmt.Sprintf(`{"query":"query($after:String$first:Int!$name:String!$owner:String!$states:[PullRequestState!]!){repository(owner: $owner, name: $name){pullRequests(first: $first, after: $after, states: $states){nodes{id,commits(last: 1){nodes{commit{committedDate},pullRequest{number,url}}},baseRefName,headRefName},pageInfo{endCursor,hasNextPage}}}}","variables":{"after":null,"first":100,"name":"%v","owner":"%v","states":["OPEN"]}}`, repo, owner))
		writeBody(t, w, fmt.Sprintf(`{"data":{"repository":{"pullRequests":{"nodes":[{"id":"007","commits":{"nodes":[{"commit":{"committedDate":"%v"},"pullRequest":{"number":7,"url":"%v"}}]},"baseRefName":"%v","headRefName":"%v"}]}}}}`, t0, url, baseName, headName))
	})
	{
		t.Run("list-prs-valid-match", func(ti *testing.T) {
			prs, err := ghApi.ListPRs(context.Background(), []string{"OPEN"}, owner, repo)
			assert.NoError(ti, err)
			assert.Len(ti, prs, 1)

			expected := &api.GitHubPR{
				Id:             "007",
				Source:         headName,
				Target:         baseName,
				LastCommitDate: t0p,
				Number:         7,
				Url:            url,
			}
			assert.Equal(ti, expected, prs[0])
		})
		t.Run("list-prs-valid-mismatch", func(ti *testing.T) {
			prs, err := ghApi.ListPRs(context.Background(), []string{"OPEN"}, owner, repo)
			assert.NoError(ti, err)
			assert.Len(ti, prs, 1)

			expected := &api.GitHubPR{
				Id:             "006",
				Source:         headName,
				Target:         baseName,
				LastCommitDate: t0p,
				Number:         7,
				Url:            url,
			}
			assert.NotEqual(ti, expected, prs[0])
		})
		t.Run("list-prs-invalid-owner", func(ti *testing.T) {
			owner = ""
			prs, err := ghApi.ListPRs(context.Background(), []string{"OPEN"}, owner, repo)
			assert.Error(ti, err)
			assert.Nil(ti, prs)
		})
		t.Run("list-prs-invalid-repo", func(ti *testing.T) {
			repo = ""
			owner = "x"
			prs, err := ghApi.ListPRs(context.Background(), []string{"OPEN"}, owner, repo)
			assert.Error(ti, err)
			assert.Nil(ti, prs)
		})
	}
	assert.NoError(t, os.Setenv(envKey, old))
}

func TestGitHub_DeleteRefs(t *testing.T) {
	envKey := "GITHUB_TOKEN"
	old := os.Getenv(envKey)
	assert.NoError(t, os.Setenv(envKey, "XXX"))

	setup(t)
	ref := "x"

	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t,
			readBody(t, r),
			fmt.Sprintf(`{"query":"mutation($input:ID!){deleteRef(input: {refId: $input}){typename :__typename}}","variables":{"input":"%v"}}`, ref))
		writeBody(t, w, `{"data":{}}`)
	})
	{
		t.Run("delete-refs-valid", func(ti *testing.T) {
			assert.NoError(ti,
				ghApi.DeleteRefs(context.Background(), ref))
		})
		t.Run("delete-refs-invalid-empty", func(ti *testing.T) {
			assert.Error(ti,
				ghApi.DeleteRefs(context.Background()))
		})
	}
	setup(t)
	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
	})
	{
		t.Run("delete-refs-errors", func(ti *testing.T) {
			err := ghApi.DeleteRefs(context.Background(), "x", "y", "z", "w")
			assert.Error(ti, err)
			assert.ErrorContains(ti, err, "unable to delete ref: x")
			assert.ErrorContains(ti, err, "unable to delete ref: y")
			assert.ErrorContains(ti, err, "unable to delete ref: z")
			assert.ErrorContains(ti, err, "unable to delete ref: w")
		})
	}
	assert.NoError(t, os.Setenv(envKey, old))
}

func TestGitHub_ClosePRs(t *testing.T) {
	envKey := "GITHUB_TOKEN"
	old := os.Getenv(envKey)
	assert.NoError(t, os.Setenv(envKey, "XXX"))

	setup(t)
	identifier := "x"

	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t,
			readBody(t, r),
			fmt.Sprintf(`{"query":"mutation($input:ID!){closePullRequest(input: {pullRequestId: $input}){typename :__typename}}","variables":{"input":"%v"}}`, identifier))
		writeBody(t, w, `{"data":{}}`)
	})
	{
		t.Run("close-prs-valid", func(ti *testing.T) {
			assert.NoError(ti,
				ghApi.ClosePRs(context.Background(), identifier))
		})
		t.Run("close-prs-invalid-empty", func(ti *testing.T) {
			assert.Error(ti,
				ghApi.ClosePRs(context.Background()))
		})
	}
	setup(t)
	handler(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
	})
	{
		t.Run("close-prs-errors", func(ti *testing.T) {
			err := ghApi.ClosePRs(context.Background(), "x", "y", "z", "w")
			assert.Error(ti, err)
			assert.ErrorContains(ti, err, "unable to close PR: x")
			assert.ErrorContains(ti, err, "unable to close PR: y")
			assert.ErrorContains(ti, err, "unable to close PR: z")
			assert.ErrorContains(ti, err, "unable to close PR: w")
		})
	}
	assert.NoError(t, os.Setenv(envKey, old))
}

/********************************/

var (
	mux   *http.ServeMux
	ghApi *api.GitHub
)

func setup(t *testing.T) {
	mux = http.NewServeMux()
	inst, err := api.NewSession(
		api.WithContext(context.Background()),
		api.WithHttpClient(&http.Client{Transport: &httpTestServer{handler: mux}}))
	assert.NoError(t, err)
	assert.NotNil(t, inst)
	ghApi = inst
}

func handler(fn http.HandlerFunc) {
	mux.HandleFunc("/graphql", fn)
}

type httpTestServer struct {
	handler http.Handler
}

func (l *httpTestServer) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	w.Header().Add("X-RateLimit-Limit", "15000")
	w.Header().Add("X-RateLimit-Used", "0")
	w.Header().Add("X-RateLimit-Remaining", "15000")
	w.Header().Add("X-RateLimit-Reset", "1693646702")
	return w.Result(), nil
}

func readBody(t *testing.T, r *http.Request) string {
	assert.NotNil(t, r)
	assert.NotNil(t, r.Body)
	body, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
	return strings.TrimSuffix(string(body), "\n")
}

func writeBody(t *testing.T, w http.ResponseWriter, body string) {
	assert.NotNil(t, w)
	n, err := io.WriteString(w, body)
	assert.NoError(t, err)
	assert.NotZero(t, n)
}

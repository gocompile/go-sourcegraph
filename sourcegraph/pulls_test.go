package sourcegraph

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/sourcegraph/go-github/github"

	"strings"

	"github.com/kr/pretty"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

func TestPullRequestsService_Get(t *testing.T) {
	setup()
	defer teardown()

	want := &PullRequest{PullRequest: github.PullRequest{Number: github.Int(1)}}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoPullRequest, map[string]string{"RepoSpec": "r.com/x", "Pull": "1"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	pull, _, err := client.PullRequests.Get(PullRequestSpec{Repo: RepoSpec{URI: "r.com/x"}, Number: 1}, nil)
	if err != nil {
		t.Errorf("PullRequests.Get returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.Get returned %+v, want %+v", pull, want)
	}
}

func TestPullRequestsService_ListByRepository(t *testing.T) {
	setup()
	defer teardown()

	want := []*PullRequest{&PullRequest{PullRequest: github.PullRequest{Number: github.Int(1)}}}
	repoSpec := RepoSpec{URI: "x.com/r"}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoPullRequests, repoSpec.RouteVars()), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"PerPage": "1",
			"Page":    "2",
		})

		writeJSON(w, want)
	})

	pulls, _, err := client.PullRequests.ListByRepository(
		repoSpec,
		&PullRequestListOptions{
			ListOptions: ListOptions{PerPage: 1, Page: 2},
		},
	)
	if err != nil {
		t.Errorf("PullRequests.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(pulls, want) {
		t.Errorf("PullRequests.List returned %+v, want %+v with diff: %s", pulls, want, strings.Join(pretty.Diff(want, pulls), "\n"))
	}
}

func TestPullRequestsService_ListComments(t *testing.T) {
	setup()
	defer teardown()

	want := []*PullRequestComment{&PullRequestComment{PullRequestComment: github.PullRequestComment{ID: github.Int(1)}}}
	pullSpec := PullRequestSpec{Repo: RepoSpec{URI: "r.com/x"}, Number: 1}

	var called bool
	mux.HandleFunc(urlPath(t, router.RepoPullRequestComments, pullSpec.RouteVars()), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"PerPage": "1",
			"Page":    "2",
		})

		writeJSON(w, want)
	})

	comments, _, err := client.PullRequests.ListComments(
		pullSpec,
		&PullRequestListCommentsOptions{
			ListOptions: ListOptions{PerPage: 1, Page: 2},
		},
	)
	if err != nil {
		t.Errorf("PullRequests.List returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(comments, want) {
		t.Errorf("PullRequests.List returned %+v, want %+v with diff: %s", comments, want, strings.Join(pretty.Diff(want, comments), "\n"))
	}
}

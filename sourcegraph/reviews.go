package sourcegraph

import (
	"strconv"
	"time"

	"github.com/sourcegraph/go-github/github"
	"github.com/sourcegraph/go-nnz/nnz"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/go-sourcegraph/router"
)

// ReviewsService communicates with the code review-related endpoints
// in the Sourcegraph API.
type ReviewsService interface {
	ListTasks(rv ReviewSpec, opt *ReviewListTasksOptions) ([]*ReviewTask, Response, error)

	ListTasksByRepo(repo RepoSpec, opt *ReviewListTasksByRepoOptions) ([]*ReviewTask, Response, error)

	ListTasksByUser(user UserSpec, opt *ReviewListTasksByUserOptions) ([]*ReviewTask, Response, error)
}

// reviewsService implements ReviewsService.
type reviewsService struct {
	client *Client
}

var _ ReviewsService = &reviewsService{}

// ReviewSpec specifies a code review.
type ReviewSpec struct {
	Repo RepoSpec // the base repository of the code review

	Number int // Sequence number of the code review
}

// RouteVars returns the route variables for generating code review
// URLs.
func (s ReviewSpec) RouteVars() map[string]string {
	return map[string]string{"RepoSpec": s.Repo.URI, "Review": strconv.Itoa(s.Number)}
}

// IssueSpec returns a specifier for the issue associated with this
// code review (same repo, same number).
func (s ReviewSpec) IssueSpec() IssueSpec {
	return IssueSpec{Repo: s.Repo, Number: s.Number}
}

// UnmarshalReviewSpec parses route variables (a map returned by
// (ReviewSpec).RouteVars()) to construct a ReviewSpec.
func UnmarshalReviewSpec(v map[string]string) (ReviewSpec, error) {
	ps := ReviewSpec{}
	var err error
	ps.Repo, err = UnmarshalRepoSpec(v)
	if err != nil {
		return ps, err
	}

	ps.Number, err = strconv.Atoi(v["Review"])
	return ps, err
}

// A ReviewTask is a task associated with a code review.
type ReviewTask struct {
	// ReviewSpec is the ReviewSpec of the code review that this task
	// is associated with.
	ReviewSpec ReviewSpec

	// Delta is the DeltaSpec for the exact base/head commit IDs of
	// the delta that this task was originally created for.
	//
	// TODO(sqs): design/explain how we determine when tasks are stale
	// when the underlying head commit changes.
	DeltaSpec DeltaSpec

	// Type is the type of review task this is. See ReviewTaskType
	// constants for the list of possible values.
	Type ReviewTaskType

	// AssigneeUID is the UID of the user to whom this task is
	// assigned, or 0 if it is not assigned to any user.
	AssigneeUID nnz.Int

	// Closed is whether this task is closed. If not, it is considered
	// to be "open."
	Closed bool

	// CreatedAt is when this task was created or generated (either
	// automatically or because of a user action).
	CreatedAt time.Time

	// The following fields are specific to this review task's type.

	// Hunk is the file and hunk header info for a
	// DiffHunkReviewTask. The FileDiff only contains one entry in its
	// Hunks array, and the Hunk's Body is empty.
	Hunk *diff.FileDiff `json:",omitempty"`

	// Def is the def that was added/changed/deleted.
	DefDelta *DefDelta `json:",omitempty"`

	// PullRequestComment is the PR comment for comment and checkbox tasks.
	PullRequestComment *PullRequestComment `json:",omitempty"`

	// IssueComment is the PR comment for comment and checkbox tasks.
	IssueComment *IssueComment `json:",omitempty"`

	// ChecklistItem is the text next to a checkbox in a comment. For
	// checklist item tasks, either PullRequestComment or IssueComment
	// is filled in as well.
	ChecklistItem string `json:",omitempty"`

	// AffectedRef is a ref to a def that this review's delta changes
	// or deletes.
	AffectedRef *DeltaAffectedRef `json:",omitempty"`

	// ExternalStatus is a commit/ref status for this review's delta's
	// head commit.
	ExternalStatus *github.RepoStatus `json:",omitempty"`
}

type ReviewTaskType string

const (
	DiffHunkReviewTask      ReviewTaskType = "diff-hunk"      // approving diff hunks
	DefReviewTask                          = "def"            // approving added/changed/deleted defs
	CommentReviewTask                      = "comment"        // resolving comments
	ChecklistItemReviewTask                = "checklist-item" // resolving checklist items
	AffectedRefReviewTask                  = "affected-ref"   // approving usages of changed/deleted defs
	ExternalReviewTask                     = "external"       // from external services (e.g., CI, coverage)
)

type ReviewListTasksCommonOptions struct {
	State string `url:",omitempty"` // "open", "closed", or "all"
}

type ReviewListTasksOptions struct {
	ReviewListTasksCommonOptions
	ListOptions
}

func (s *reviewsService) ListTasks(rv ReviewSpec, opt *ReviewListTasksOptions) ([]*ReviewTask, Response, error) {
	url, err := s.client.URL(router.ReviewTasks, rv.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*ReviewTask
	resp, err := s.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, nil
}

type ReviewListTasksByRepoOptions struct {
	ReviewListTasksCommonOptions

	// The User login for whom to fetch tasks (usually the currently
	// authenticated user).
	User string `url:",omitempty"`

	ListOptions
}

func (s *reviewsService) ListTasksByRepo(repo RepoSpec, opt *ReviewListTasksByRepoOptions) ([]*ReviewTask, Response, error) {
	url, err := s.client.URL(router.RepoReviewTasks, repo.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*ReviewTask
	resp, err := s.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, nil
}

type ReviewListTasksByUserOptions struct {
	ReviewListTasksCommonOptions
	ListOptions
}

func (s *reviewsService) ListTasksByUser(user UserSpec, opt *ReviewListTasksByUserOptions) ([]*ReviewTask, Response, error) {
	url, err := s.client.URL(router.UserReviewTasks, user.RouteVars(), opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*ReviewTask
	resp, err := s.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, nil
}

var _ ReviewsService = &MockReviewsService{}

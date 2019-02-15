package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"github.com/shurcooL/githubv4" // Serious concern that v4 is an inferior choice to v3
	"golang.org/x/oauth2"
)

var (
	ErrBadRepo = errors.New("poorly formatted repo name")
)

// Issue structure represents a GitHub issue object and a portion of fields available.
// NOTE: This structure is declared by GitHub.com
type Issue struct {
	// Title is the title of the GitHub issue
	Title string
	// Repository contains fields describing repo the issue is on
	Repository struct {
		// Name is the issue's repository's Name
		Name string
		// Owner is an object describing the user who owns the repository described by the repository object
		Owner struct {
			// Login is the login name of the issue's repository's owner
			Login string
		}
	}
	// Body is the body of the issue
	Body string
	// Author is an object containing information about the issue's author
	Author struct {
		// Login is the login name of the author object
		Login string
	}
	// Number is the issue number
	Number int
	// Url is the issue's url
	Url string
}

// NOTE graphQL basics-
// graphQL Queries : SQL Select :: graphQL Mutations : SQL Upsert
// 1) Create an input structure (for mutations) or
// variables map (for queries) to supply arguments.
// 2) Create a struct that reflects path to object and return fields. Eg: Issue

// GitHubIssueBot is a helper type essentially wrapping oauth2 and githubv4.
type GitHubIssueBot struct {
	client     *githubv4.Client
	httpClient *http.Client
	token      string
}

// NOTE: The following two declarations are used to enable "preview mode" in github v4 API.

// Transport is a oauth2.Transport wrapper (an http.Transport wrapper itself)
// enabling us to add headers to all requests.
type Transport struct {
	http.RoundTripper
}

// RoundTrip is a wrapper over oauth2.Transport.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// This header is required by github for creating issues.
	req.Header.Add("Accept", `application/vnd.github.starfire-preview+json`)
	return t.RoundTripper.RoundTrip(req)
}

// NewGitHubIssueBot returns a GitHubIssueBot with it's token set.
func NewGitHubIssueBot(ctx context.Context, token string) *GitHubIssueBot {
	gBot := &GitHubIssueBot{
		token: token,
	}
	// NOTE: Connect is split into a seperate function so that it can be used to reconnect if needed.
	gBot.Connect(ctx)
	return gBot
}

// Connect creates an http.Client with oauth2 and attempts to connect to GitHub.
func (g *GitHubIssueBot) Connect(ctx context.Context) {
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	g.httpClient = oauth2.NewClient(ctx, tokenSrc)

	// We're wrapping the RoundTripper oath2 just gave us.
	newTransport := &Transport{RoundTripper: g.httpClient.Transport}
	// We're giving oath2 the wrapped RoundTripper.
	g.httpClient.Transport = newTransport

	g.client = githubv4.NewClient(g.httpClient)
	log.Infof("IssueBot connected to github")
	return
}

// CheckToken finds the full name of an organization based off the "URL" name.
func (g *GitHubIssueBot) CheckToken(ctx context.Context) (name string, login string, err error) {

	var query struct {
		Viewer struct {
			Login string
			Name  string
		}
	}

	// Try a different query before returning err...
	err = g.client.Query(ctx, &query, nil)
	if err != nil {
		return "", "", trace.Wrap(err)
	}
	return query.Viewer.Name, query.Viewer.Login, nil
}

// NewIssue takes a repo, issue, and issueBody and then creates a new issue.
func (g *GitHubIssueBot) NewIssue(ctx context.Context, repo string, title string, body string) (*Issue, error) {
	repoPath := strings.Split(repo, "/")
	if len(repoPath) != 2 {
		// TODO: check channel for reponame
		return nil, ErrBadRepo
	}
	// We need to see if the repo exists first. Search would still be better.
	variables := map[string]interface{}{
		"org":  githubv4.String(repoPath[0]),
		"repo": githubv4.String(repoPath[1]),
	}

	var query struct {
		Repository struct {
			ID githubv4.ID
		} `graphql:"repository(name: $repo, owner: $org)"`
	}

	if err := g.client.Query(ctx, &query, variables); err != nil {
		return nil, trace.Wrap(err)
	}

	// NOTE: This type should eventually be provided by the GitHubV4 dependency
	// NOTE: pkg githubv4 depends on this type name
	type CreateIssueInput struct {
		Title            githubv4.String  `json:"title"`
		Body             githubv4.String  `json:"body"`
		RepositoryId     githubv4.ID      `json:"repositoryId"`
		ClientMutationID *githubv4.String `json:"clientMutationId,omitempty"`
	}

	input := CreateIssueInput{
		Title:        githubv4.String(title),
		Body:         githubv4.String(body),
		RepositoryId: query.Repository.ID,
	}

	var m struct {
		CreateIssue struct {
			Issue Issue // TODO: This seems awkward
		} `graphql:"createIssue(input: $input)"`
	}

	if err := g.client.Mutate(ctx, &m, input, nil); err != nil {
		return nil, trace.Wrap(err)
	}

	return &m.CreateIssue.Issue, nil
}

package main

import (
	"context"
	"net/http"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"github.com/shurcooL/githubv4" // Serious concern that v4 is an inferior choice to v3
	"golang.org/x/oauth2"
)

// Issue structure represents a GitHub issue object and a portion of fields available. NOTE: This structure is declared by GitHub.com
type Issue struct {
	Title      string
	Repository struct {
		Name  string
		Owner struct {
			Login string
		}
	}
	Body   string
	Author struct {
		Login string
	}
	Number int
	Url    string
}

// Helpful notes for GraphQL queries:
// 1) Create an input structure or variables map to supply arguments.
// 2) Create a structure that reflects path to object and desired return fields. Eg: Issue
// It's particular about capitalizing or not

// GitHubIssueBot is a helper type to initialize and call common functions easily
type GitHubIssueBot struct {
	client     *githubv4.Client
	httpClient *http.Client
	token      string
	org        string
}

// NOTE: The following two declerations are used to enable "preview mode" in github v4 API

// Transport is a oauth2.Transport wrapper (a http.Transport wrapper itself ) enabling us to add headers to all requests.
type Transport struct {
	http.RoundTripper
}

// RoundTrip is a wrapper over oauth2.Transport.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// This header is required by github for creating issues
	req.Header.Add("Accept", `application/vnd.github.starfire-preview+json`)
	return t.RoundTripper.RoundTrip(req)
}

// NewGitHubIssueBot returns a GitHubIssueBot with it's token set.
func NewGitHubIssueBot(ctx context.Context, token string) *GitHubIssueBot {
	gBot := &GitHubIssueBot{
		token: token,
	}
	// Note: Connect is split into a seperate function so that it can be used to reconnect if needed
	gBot.Connect(ctx)
	return gBot
}

// Connect creates an http.Client with oauth2 and attempts to connect to GitHub
func (g *GitHubIssueBot) Connect(ctx context.Context) (err error) {
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	g.httpClient = oauth2.NewClient(ctx, tokenSrc)

	// We're wrapping the RoundTripper oath2 just gave us
	newTransport := &Transport{RoundTripper: g.httpClient.Transport}
	// We're giving oath2 the wrapped RoundTripper
	g.httpClient.Transport = newTransport

	g.client = githubv4.NewClient(g.httpClient)
	log.Infof("IssueBot connected to github")
	return
}

// SetToken is just a setter for the private token member of the type
func (g *GitHubIssueBot) SetToken(token string) {
	g.token = token
}

// GetOrg() is a getter for the specified organziation's full name
func (g *GitHubIssueBot) GetOrg() string {
	return g.org
}

// CheckOrg finds the full name of an organization based off the "URL" name
func (g *GitHubIssueBot) CheckOrg(ctx context.Context, org string) error {

	variables := map[string]interface{}{
		"org": githubv4.String(org),
	}

	var queryUser struct { // TODO: use Search object + type, not User/Org object
		User struct {
			Name string
		} `graphql:"user(login: $org)"`
	}

	var queryOrg struct {
		Organization struct {
			Name string
		} `graphql:"organization(login: $org)"`
	}

	// TODO: This logic will simplify when TODO above is addressed
	if err := g.client.Query(ctx, &queryUser, variables); err != nil {
		// Try a different query before returning err...
		if err = g.client.Query(ctx, &queryOrg, variables); err != nil {
			return trace.Wrap(err)
		}
	}
	g.org = org
	return nil
}

// NewIssue takes a repo, issue, and issueBody and creates a new issue
func (g *GitHubIssueBot) NewIssue(ctx context.Context, repo string, title string, body string) (*Issue, error) { // TODO TODO TODO ISSUE CONFIG

	// We need to see if the repo exists first. Search would still be better.
	variables := map[string]interface{}{
		"org":  githubv4.String(g.org),
		"repo": githubv4.String(repo),
	}

	var query struct {
		Repository struct {
			ID githubv4.ID
		} `graphql:"repository(name: $repo, owner: $org)"`
	}

	if err := g.client.Query(ctx, &query, variables); err != nil {
		return nil, trace.Wrap(err)
	}

	// Preparing some types for a "mutate" query
	// This type should eventually be provided by the GitHubV4 dependency
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
			Issue Issue // TODO: I really don't know what to do about this. I like _t...
		} `graphql:"createIssue(input: $input)"`
	}

	if err := g.client.Mutate(ctx, &m, input, nil); err != nil {
		return nil, trace.Wrap(err)
	}

	return &m.CreateIssue.Issue, nil
}

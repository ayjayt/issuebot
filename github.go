package main

import (
	"context"
	"errors"
	"github.com/mailgun/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"net/http"
)

var (
	errNoOrg = errors.New("Either user doesn't have access or org doesn't exist")
)

// GitHubIssuebot is a helper type so we can initialize and call common functions easily
// TODO: can we do to this what gravitational/hello did for Helloer?
type GitHubIssueBot struct {
	client     *githubv4.Client
	httpClient *http.Client
	token      string
	org        string
}

// NewGitHubIssueBot is basically a tiny initializer
func NewGitHubIssueBot(token string) (bot *GitHubIssueBot) {
	// Really bothers me that this declares a pointer AND the value it points to- creates heap
	bot = &GitHubIssueBot{
		token: token,
	}
	return bot
}

// GetOrg() is a getter for the registered org
func (g *GitHubIssueBot) GetOrg() (org string) {
	return org
}

// CheckOrg sanity checks that we can access the organization passed as a parameter. How do I feel about v4? I dont' know.
func (g *GitHubIssueBot) CheckOrg(org string) (ok bool, err error) { // must paginate, repository

	variables := map[string]interface{}{
		"org": githubv4.String(org),
	}

	var name string

	var queryUser struct { // TODO: user search + type, not User/Org
		User struct {
			Name string
		} `graphql:"user(login: $org)"`
	}

	var queryOrg struct {
		Organization struct {
			Name string
		} `graphql:"organization(login: $org)"`
	}
	// BUG(AJ) So hacky I'll call it a bug. Infact, error handling throughout everything here is crap. Not sure if teams can own repos either.

	err = g.client.Query(context.Background(), &queryUser, variables)
	if err != nil {
		err = g.client.Query(context.Background(), &queryOrg, variables)
		if err == nil {
			name = queryOrg.Organization.Name
		}
	} else {
		name = queryUser.User.Name
	}
	if err != nil {
		log.Errorf("Error in CheckOrg() couldn't access supplied org: %T: %v", err, err) // BUG(AJ) GitHub is kicking back the auth token sometimes, e.g. LOG: -LEGIT TOKEN-\n
		return false, err
	}
	log.Debugf("Display Name of %v: %v", org, name)
	g.org = org
	return true, nil

}

// NewIssue takes a repo, issue, and issueBody and creates a new issue
func (g *GitHubIssueBot) NewIssue(repo string, title string, body string) (URL string, err error) {
	variables := map[string]interface{}{
		"org":  githubv4.String(g.org),
		"repo": githubv4.String(repo),
	}

	var query struct {
		Repository struct {
			ID githubv4.ID
		} `graphql:"repository(name: $repo, owner: $org)"`
	}

	err = g.client.Query(context.Background(), &query, variables)

	if err != nil {
		log.Errorf("Error in NewIssue() on query: %T: %v", err, err)
		return "", err
	}

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
			Issue struct {
				Url string
			}
		} `graphql:"createIssue(input: $input)"`
	}
	err = g.client.Mutate(context.Background(), &m, input, nil)
	if err != nil {
		log.Errorf("Error in NewIssue() on mutate: %T: %v", err, err)
		return "", err
	}
	return string(m.CreateIssue.Issue.Url), nil

	// TODO: so much other stuff that should go along with posting the issue- assigning it, etc
}

// Connect just replaces the internal variabels of the GitHubIssueBot structure- so it can be used to reconnect
// TODO: It should be Disconnect/Reconnecting if not nil but this is okay for now
func (g *GitHubIssueBot) Connect() (err error) {
	tokenSrc := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	g.httpClient = oauth2.NewClient(context.Background(), tokenSrc)

	// We're wrapping the transport oath2 just gave us with ours from below
	// but ours is really just a wrapper for the one we just got
	newTransport := &Transport{RoundTripper: g.httpClient.Transport}
	g.httpClient.Transport = newTransport

	g.client = githubv4.NewClient(g.httpClient)
	log.Infof("IssueBot connected to github")
	return nil
	//TODO: What errors should be here?
}

// SetToken is just a setter for the private token member of the type
func (g *GitHubIssueBot) SetToken(token string) {
	g.token = token
}

// Transport is a oauth2.Transport wrapper (a http.Transport wrapper itself ) enabling us to add headers to all requests.
type Transport struct {
	http.RoundTripper
}

// RoundTrip is a wrapper over oauth2.Transport.RoundTrip.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// This header is required by github for creating issues
	req.Header.Add("Accept", `application/vnd.github.starfire-preview+json`)
	return t.RoundTripper.RoundTrip(req)
}

// TODO: this Transport wrapper would be useful in more contexts (including eliminating middleware in http server)
// if it had an initialization function that checked if the received Transport was nil and if so making sure
// to initalize the http.DefaultTransport instead of 1) calling a non-function in RoundTrip and 2) tricking pkg/http which checks
// if Transport == nil before calling http.DefaultTransport in http.Client.Do()

/*
	Personal Note: This is without a doubt, the Transport wrap, the wildest thing I've done.
	This is considering how much of a mess the oauth2/http implementation is ("Transport" is a variable name AND a type name)
	And the mess that is oauth's Transport wrapper over http's Transport makes my eyes bleed.
	http.Client is a structure with a Transport member of interface-type RoundTripper, meaning it must implement RoundTrip().
	Usually http.Client uses it's http.Transport type as a default.
	oauth2.NewClient returns a http.Client which uses it's own Transport type, which is a structure that a) is a http.RoundTripper,
	but b) also contains the original http.Transport as a `base http.Transport` member which it calls in it's own RoundTrip()
	after doing this weird *http.Request mashup - AJ 2018/02/08
*/

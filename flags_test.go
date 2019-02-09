package main

import (
	"github.com/gravitational/trace"
	. "gopkg.in/check.v1"
)

// Check package needs a type
type FlagsSuite struct{}

var _ = Suite(&FlagsSuite{})

func (s *FlagsSuite) SetUpSuite(c *C) {
	// This may be used as test coverage improves
}

func (s *FlagsSuite) SetUpTest(c *C) {
	flagOrg = nil
	flagAuth = nil
	flagSlackToken = nil
	flagSlackTokenFile = nil
	flagGitHubToken = nil
	flagGitHubTokenFile = nil
}

func (s *FlagsSuite) TearDownTest(C *C) {
	// This may be used as test coverage improves
}

// TODO: comments
func (s *FlagsSuite) TestVerifyFlags(c *C) {
	var nilError error = nil
	testTables := []struct {
		name            string
		org             string
		auth            string
		slackToken      string
		slackTokenFile  string
		githubToken     string
		githubTokenFile string
		expected        error
	}{
		{name: "All Errors", org: "", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "fake-test-slack-filename", githubToken: "fake-test-github-token", githubTokenFile: "fake-github-filename", expected: ErrBadFlag},
		{name: "No Org", org: "", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "", githubToken: "fake-test-github-token", githubTokenFile: "", expected: ErrBadFlag},
		{name: "All Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "fake-test-slack-filename", githubToken: "fake-test-github-token", githubTokenFile: "fake-github-filename", expected: ErrBadFlag},
		{name: "GitHub Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "", githubToken: "fake-test-github-token", githubTokenFile: "fake-github-filename", expected: ErrBadFlag},
		{name: "Slack Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "fake-test-slack-filename", githubToken: "", githubTokenFile: "fake-github-filename", expected: ErrBadFlag},
		{name: "All Defaults", org: "fake-org-name", auth: "", slackToken: "", slackTokenFile: "", githubToken: "", githubTokenFile: "", expected: nilError},
		{name: "Best Use w/ Default Auth", org: "fake-org-name", auth: "", slackToken: "", slackTokenFile: "fake-test-slack-filename", githubToken: "", githubTokenFile: "fake-github-filename", expected: nilError},
		{name: "Best Use w/ Custom Auth", org: "fake-org-name", auth: "fake-auth-filename", slackToken: "", slackTokenFile: "fake-test-slack-filename", githubToken: "", githubTokenFile: "fake-github-filename", expected: nilError},
		{name: "Token Overrides", org: "fake-org-name", auth: "fake-auth-filename", slackToken: "fake-slack-token", slackTokenFile: "", githubToken: "fake-test-github-token", githubTokenFile: "", expected: nilError},
	}
	for i, testTable := range testTables {
		// Setting up comments to go alog with failures
		comment := Commentf("test #%d (%v)- too many arguments to print", i+1, testTable.name)

		flagOrg = &testTable.org
		flagAuth = &testTable.auth
		flagSlackToken = &testTable.slackToken
		flagSlackTokenFile = &testTable.slackTokenFile
		flagGitHubToken = &testTable.githubToken
		flagGitHubTokenFile = &testTable.githubTokenFile

		err := verifyFlagsSanity()

		c.Assert(err, Equals, testTable.expected, comment)
	}
}

func (s *FlagsSuite) TestReadAuth(c *C) {
	// create a temporary auth file
	// create auth flag
	// check against slice
	c.Skip("Skipping this to move on, but should come back")
}
func (s *FlagsSuite) TestReadGitHubToken(c *C) {
	// create a temporary auth file
	// create auth flag
	// check against string
	// also test with flag token
	c.Skip("Skipping this to move on, but should come back")
}
func (s *FlagsSuite) TestReadSlackToken(c *C) {
	// create a temporary auth file
	// create auth flag
	// check against string
	// also taken with flag token
	c.Skip("Skipping this to move on, but should come back")
}

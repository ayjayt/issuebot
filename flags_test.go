package main

import (
	. "gopkg.in/check.v1" // not sure why we went this route (with .) but I'm taking hints from gravitational/hello
)

// Check package needs a type
type FlagsSuite struct{}

var _ = Suite(&FlagsSuite{})

func (s *FlagsSuite) SetUpSuite(c *C) {
	// If you want
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
	// If you want
}

// TODO: comments
func (s *FlagsSuite) TestVerifyFlags(c *C) {
	var nilError error = nil
	testTables := []struct {
		name              string
		org               string
		auth              string
		slack_token       string
		slack_token_file  string
		github_token      string
		github_token_file string
		expected          error
	}{ // no org, token and file X2, 5x rational cases
		{name: "All Errors", org: "", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "fake-test-slack-filename", github_token: "fake-test-github-token", github_token_file: "fake-github-filename", expected: ErrBadFlag},
		{name: "No Org", org: "", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "", github_token: "fake-test-github-token", github_token_file: "", expected: ErrBadFlag},
		{name: "All Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "fake-test-slack-filename", github_token: "fake-test-github-token", github_token_file: "fake-github-filename", expected: ErrBadFlag},
		{name: "GitHub Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "", github_token: "fake-test-github-token", github_token_file: "fake-github-filename", expected: ErrBadFlag},
		{name: "Slack Token Redundancy", org: "fake-org-name", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "fake-test-slack-filename", github_token: "", github_token_file: "fake-github-filename", expected: ErrBadFlag},
		{name: "All Defaults", org: "fake-org-name", auth: "", slack_token: "", slack_token_file: "", github_token: "", github_token_file: "", expected: nilError},
		{name: "Best Use w/ Default Auth", org: "fake-org-name", auth: "", slack_token: "", slack_token_file: "fake-test-slack-filename", github_token: "", github_token_file: "fake-github-filename", expected: nilError},
		{name: "Best Use w/ Custom Auth", org: "fake-org-name", auth: "fake-auth-filename", slack_token: "", slack_token_file: "fake-test-slack-filename", github_token: "", github_token_file: "fake-github-filename", expected: nilError},
		{name: "Token Overrides", org: "fake-org-name", auth: "fake-auth-filename", slack_token: "fake-slack-token", slack_token_file: "", github_token: "fake-test-github-token", github_token_file: "", expected: nilError},
	}
	for i, testTable := range testTables {
		// setting up comments to go alog with failures
		comment := Commentf("test #%d (%v)- too many arguments to print", i+1, testTable.name)

		flagOrg = &testTable.org
		flagAuth = &testTable.auth
		flagSlackToken = &testTable.slack_token
		flagSlackTokenFile = &testTable.slack_token_file
		flagGitHubToken = &testTable.github_token
		flagGitHubTokenFile = &testTable.github_token_file

		// running actual test
		err := verifyFlagsSanity()

		// asserting outptus
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

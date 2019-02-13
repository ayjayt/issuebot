package main

import (
	"github.com/gravitational/trace"
	. "gopkg.in/check.v1"
)

// Check package needs a type
// TODO
type FlagsSuite struct{}

var _ = Suite(&FlagsSuite{})

// TODO
func (s *FlagsSuite) SetUpSuite(c *C) {
	// This may be used as test coverage improves
}

// TODO
func (s *FlagsSuite) SetUpTest(c *C) {
	// This may be used as test coverage improves
}

// TODO
func (s *FlagsSuite) TearDownTest(C *C) {
	// This may be used as test coverage improves
}

// TODO
func (s *FlagsSuite) TestPopulateFlags(c *C) {
	// NOTE: I took a shot at a comprehensive test table type. Might be a bit much.

	// configResult combines the value (via actual type) and error into a structure
	type configResult struct {
		config
		err error
	}
	// Go configReslult result example
	goodResult := configResult{
		config: config{
			slackToken:  "fake-slack-token",
			gitHubToken: "fake-github-token",
			org:         "fake-org-name",
			authFile:    "fake-auth-path",
		},
		err: nil,
	}
	// Testtable has meta data + configResult
	testTables := []struct {
		name string
		configResult
	}{
		{name: "All Errors",
			configResult: configResult{
				config: config{org: "", authFile: goodResult.authFile, slackToken: "", gitHubToken: ""},
				err:    ErrBadFlag,
			},
		},
		{name: "No Org",
			configResult: configResult{
				config: config{org: "", authFile: goodResult.authFile, slackToken: goodResult.slackToken, gitHubToken: goodResult.gitHubToken},
				err:    ErrBadFlag,
			},
		},
		{name: "No GitHub Token",
			configResult: configResult{
				config: config{org: goodResult.org, authFile: goodResult.authFile, slackToken: goodResult.slackToken, gitHubToken: ""},
				err:    ErrBadFlag,
			},
		},
		{name: "No Slack Token",
			configResult: configResult{
				config: config{org: goodResult.org, authFile: goodResult.authFile, slackToken: "", gitHubToken: ""},
				err:    ErrBadFlag,
			},
		},
		{name: "No Errors",
			configResult: configResult{
				config: config{org: goodResult.org, authFile: goodResult.authFile, slackToken: goodResult.slackToken, gitHubToken: goodResult.gitHubToken},
				err:    nil,
			},
		},
	}
	for i, tt := range testTables {

		cfg, err := populateFlags(tt.org, tt.slackToken, tt.gitHubToken, tt.authFile)
		cfgRes := configResult{config: cfg, err: trace.Unwrap(err)}

		comment := func(cfgRes configResult) CommentInterface {
			return Commentf("test #%d (%v)-\nExpect: %+v\nGotten: %+v", i+1, tt.name, tt.configResult, cfgRes)
		}
		c.Assert(tt.configResult, DeepEquals, cfgRes, comment(cfgRes))
	}
}

func (s *FlagsSuite) TestReadAuth(c *C) {
	// create a temporary auth file
	// create auth flag
	// check against slice
	c.Skip("Skipping this to move on, but should come back")
}

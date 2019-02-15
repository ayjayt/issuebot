package main

import (
	"errors"
	"flag"
	"strings"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"io/ioutil"
)

const (
	// defaultAuthFilePath is used in the flags list
	defaultAuthFilePath = "./userlist"
)

var (
	// ErrBadFlag is returned whenever the user has improperly run the program.
	ErrBadFlag = errors.New("command was run improperly, check --help")
)

// NOTE: flags can be defined anywhere, they're defined as a block here
// to see the command-line UX as a whole.
var (
	// flagOrg is the name of the github organization to access.
	flagOrg = flag.String("org",
		"",
		"Organization bot has access to")

	// flagAuth is path to textfile of authorized users.
	flagAuthFile = flag.String("auth",
		defaultAuthFilePath,
		"What file contains a list of authorized users")

	// flagSlackToken is a slack token.
	flagSlackToken = flag.String("slack_token",
		"",
		"Specify the slack token")

	// github_token is a github token.
	flagGitHubToken = flag.String("github_token",
		"",
		"Specify the github oauth token")
)

type config struct {
	slackToken  string
	gitHubToken string
	org         string
	authFile    string
	authedUsers []string
}

func init() {
	flag.Parse()
}

// flagHelper calls populateFlags with the flags above. These functions are
// seperate to allow unit testing the logic in populateFlags.
func flagHelper() (config, error) {
	c, err := populateFlags(*flagOrg, *flagSlackToken, *flagGitHubToken, *flagAuthFile)
	if err != nil {
		return c, err
	}
	err = c.loadAuthedUsers()
	return c, err
}

// populateFlags checks flag validity and initializes a "config" struct.
func populateFlags(org, slackToken, gitHubToken, authFile string) (config, error) {

	c := config{}
	// NOTE: It's more efficient (in the long run) to copy this structure by value

	// TODO: Implement an errors structure that contains an []error.
	// It must implement the "Error" interface.
	// It will have a receiver function .contains(err) to check if the error contains.

	var err error
	if len(org) == 0 {
		log.Errorf("You must specify an organization with --org")
		err = ErrBadFlag
	}
	c.org = org

	if len(slackToken) == 0 {
		log.Errorf("You must specify a Slack token with --slack_token")
		err = ErrBadFlag
	}
	c.slackToken = slackToken

	if len(gitHubToken) == 0 {
		log.Errorf("You must specify a GitHub token with --github_token")
		err = ErrBadFlag
	}
	c.gitHubToken = gitHubToken

	c.authFile = authFile

	if err != nil {
		flag.PrintDefaults()
		return c, trace.Wrap(err)
	}

	return c, trace.Wrap(err)
}

// loadAuthedUsers maps a newline deliminated list of users to a string slice.
func (c *config) loadAuthedUsers() error {
	authFileContents, err := ioutil.ReadFile(c.authFile)
	if err != nil {
		return trace.Wrap(err)
	}
	authedUsers := strings.Split(string(authFileContents), "\n")

	// NOTE: The last slice element after strings.Split is empty, so truncate
	c.authedUsers = authedUsers[:len(authedUsers)-1]
	return nil
}

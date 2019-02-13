package main

import (
	"errors"
	"flag"
	"strings"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"io/ioutil"
)

var (
	// ErrBadFlag is returned whenever the user has improperly run the program
	ErrBadFlag = errors.New("command was run improperly, check --help")
)

// Note: flags with the "flag" package can be defined anywhere, but having them defined as a block here allows the developer to focus on the command-line UX as a whole
var (
	// flagOrg is the name of the org the bot has access to
	flagOrg = flag.String("org",
		"",
		"Organization bot has access to")

	// flagAuth provides a file by which to load authorized users
	flagAuthFile = flag.String("auth",
		"./userlist",
		"What file contains a list of authorized users")

	// flagSlackToken will provide a slack token manually
	flagSlackToken = flag.String("slack_token",
		"",
		"Specify the slack token")

	// github_token will provide a github token manually
	flagGitHubToken = flag.String("github_token",
		"",
		"Specify the github oauth token")
)

type config struct {
	slackToken  string
	gitHubToken string
	org         string
	authedUsers []string
}

func init() {
	flag.Parse()
}

// flagHelper calls populateFlags with the flag variables defined above. The functions are seperate to allow unit testing the logic.
func flagHelper() (config, error) {
	return populateFlags(*flagOrg, *flagSlackToken, *flagGitHubToken, *flagAuthFile)
}

// populateFlags is an initializer which does a basic check on the flags and then defines "config" structure memembers
func populateFlags(org string, slackToken string, gitHubToken string, authFile string) (config, error) {

	c := config{} // It's more efficient (in the long run) to copy this structure by value

	// TODO: Implement an errors structure that contains an []error.
	// It must implement the "Error" interface.
	// It will have a receiver function .contains(err) to check if the error contains.

	// Now do a basic sanity test on flags. Run through all tests and inform user completely before returning.
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

	if err != nil {
		flag.PrintDefaults()
		return c, trace.Wrap(err)
	}

	// Load authorized users from file
	c.authedUsers, err = loadAuthedUsers(authFile)
	return c, trace.Wrap(err)
}

// loadAuthedUsers maps a new-line deliminated list of users to a string slice 
func loadAuthedUsers(authFile string) ([]string, error) {
	authFileContents, err := ioutil.ReadFile(authFile)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	authedUsers := strings.Split(string(authFileContents), "\n")
	return authedUsers[:len(authedUsers)-1], nil // Don't return the final empty newline characteristic of strings.Split()
}

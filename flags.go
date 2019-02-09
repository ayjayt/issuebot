package main

import (
	"errors"
	"flag"
	"strings"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"io/ioutil"
)

// default paths for issuebot to look for auth tokens
var (
	defaultSlackTokenFile  = "./slack_token"
	defaultGitHubTokenFile = "./github_token"
)
var (
	// ErrBadFlag is returned whenever the user has improperly run the program
	ErrBadFlag = errors.New("command was run improperly, check --help")
)
var (
	// flagOrg is the name of the org the bot has access to
	flagOrg = flag.String("org",
		"",
		"Organization bot has access to")

	// flagAuth provides a file by which to load authorized users
	flagAuth = flag.String("auth",
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

	// flagSlackTokenFile will provide a filename for a slack token
	flagSlackTokenFile = flag.String("slack_token_file",
		"",
		"Specify the slack token file")

	// github_token_file will provide a filename for a github token
	flagGitHubTokenFile = flag.String("github_token_file",
		"",
		"Specify the github oauth token file")
)

type config struct {
	slackToken  string
	githubToken string
	org         string
	authedUsers []string
}

// checkAndSetDefautls parses the flags, does a basic check on sanity, and then "processes" them
func checkAndSetDefaults() (c config, err error) {
	// Read the flags in
	flag.Parse()

	// Now do a basic sanity test on flags.
	err = verifyFlagsSanity()
	return c, trace.Wrap(err)

	c.org = *flagOrg

	// Prepare GitHub
	c.githubToken, err = loadGitHubToken() // flags.go
	if err != nil {
		return c, err
	}

	// Prepare Slack
	c.slackToken, err = loadSlackToken() // flags.go
	if err != nil {
		return c, err
	}
	c.authedUsers, err = loadAuthedUsers()
	if err != nil {
		return c, err
	}
	return c, nil
}

// verifyFlagsSanity just does a basic check on provided flags- are the ones that need to be there, there?
// This function reads from globals (flags)
func verifyFlagsSanity() error {
	var err error
	if len(*flagOrg) == 0 {
		log.Errorf("You must specify an organization with --org")
		err = ErrBadFlag
	}
	if *flagSlackTokenFile == "" && *flagSlackToken == "" {
		flagSlackTokenFile = &defaultSlackTokenFile
	} else if *flagSlackTokenFile != "" && *flagSlackToken != "" {
		log.Errorf("You must not specify both --flagSlackTokenFile AND --flagSlackToken")
		err = ErrBadFlag
	}
	if *flagGitHubTokenFile == "" && *flagGitHubToken == "" {
		flagGitHubTokenFile = &defaultGitHubTokenFile
	} else if *flagGitHubTokenFile != "" && *flagGitHubToken != "" {
		log.Errorf("You must not specify both --flagGitHubTokenFile AND --flagGitHubToken")
		err = ErrBadFlag
	}
	if err != nil {
		flag.PrintDefaults()
	}
	return trace.Wrap(err)
}

// loadAuthedUsers reads the flagAuth path and returns users.
// This function reads from globals (flags)
func loadAuthedUsers() (ret []string, err error) {
	authFile, err := ioutil.ReadFile(*flagAuth)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	ret = strings.Split(string(authFile), "\r\n")
	if len(ret) == 1 { // it's possible that different OSes have different newline conventions- don't check \n first
		ret = strings.Split(ret[0], "\n")
	}
	return ret[:len(ret)-1], nil
}

// loadSlackTokentries to return a slack key (from flag or file)
// This function reads from globals (flags)
func loadSlackToken() (string, error) {
	if *flagSlackToken != "" {
		return *flagSlackToken, nil
	}
	slackFileContents, err := ioutil.ReadFile(*flagSlackTokenFile)
	if err != nil {
		return "", trace.Wrap(err)
	}
	slack := string(slackFileContents)

	// BUG(AJ) tokens have text-editor artifact- I don't want to modify tokens though.
	if strings.HasSuffix(slack, "\r\n") {
		log.Warningf("Slack Token ends in whitespace, eliminating two characters (\\r\\n)...")
		slack = strings.TrimSuffix(slack, "\r\n")
	} else if strings.HasSuffix(slack, "\n") {
		log.Warningf("Slack Token ends in whitespace, eliminating one character (\\n)...")
		slack = strings.TrimSuffix(slack, "\n")
	}
	return slack, nil
}

// loadGitHubToken tries to return a github key (from flag or file)
// This function reads from globals (flags)
func loadGitHubToken() (string, error) {
	if *flagGitHubToken != "" {
		return *flagGitHubToken, nil
	}
	githubFileContents, err := ioutil.ReadFile(*flagGitHubTokenFile)
	if err != nil {
		return "", trace.Wrap(err)
	}
	github := string(githubFileContents)
	if strings.HasSuffix(github, "\r\n") {
		log.Warningf("GitHub Token ends in whitespace, eliminating two characters (\\r\\n)...")
		github = strings.TrimSuffix(github, "\r\n")
	} else if strings.HasSuffix(github, "\n") {
		log.Warningf("GitHub Token ends in whitespace, eliminating one character (\\n)...")
		github = strings.TrimSuffix(github, "\n")
	}
	return github, nil
}

package main

import (
	"errors"
	"flag"
	"github.com/mailgun/log"
	"io/ioutil"
	"strings"
)

// these vars make checking if flags were set easier than using the flag pkg default
// they can't be const because golang wants to optimize them away but we use their addresses
var (
	DefaultSlackTokenFile  = "./slack_token"
	DefaultGitHubTokenFile = "./github_token"
)
var (
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

func flagInit() (err error) {
	// Read the flags in
	flag.Parse()

	// Now do a basic sanity test on flags.
	err = verifyFlagsSanity()
	return err
}

// verifyFlagsSanity just does a basic check on provided flags- are the ones that need to be there, there?
// This function reads from globals (flags)
func verifyFlagsSanity() (err error) {
	if len(*flagOrg) == 0 {
		log.Errorf("You must specify an organization, see --help")
		err = ErrBadFlag
	}
	if *flagSlackTokenFile == "" {
		if *flagSlackToken == "" {
			flagSlackTokenFile = &DefaultSlackTokenFile
		}
	} else {
		if *flagSlackToken != "" {
			log.Errorf("You must not specify both --flagSlackTokenFile AND --flagSlackToken, see --help")
			err = ErrBadFlag
		}
	}
	if *flagGitHubTokenFile == "" {
		if *flagGitHubToken == "" {
			flagGitHubTokenFile = &DefaultGitHubTokenFile
		}
	} else {
		if *flagGitHubToken != "" {
			log.Errorf("You must not specify both --flagGitHubTokenFile AND --flagGitHubToken, see --help")
			err = ErrBadFlag
		}
	}
	return err
}

// loadAuthedUsers reads the file specified by flagAuth to create a list of authorized slack users. The caller can decide whether or not to exit on error.
// This function reads from globals (flags)
func loadAuthedUsers() (ret []string, err error) {
	var authFile []byte
	authFile, err = ioutil.ReadFile(*flagAuth)
	if err != nil {
		return nil, err
	}
	ret = strings.Split(string(authFile), "\r\n")
	if len(ret) == 1 { // it's possible that different OSes have different newline conventions- don't check \n first
		ret = strings.Split(ret[0], string(10)) // this is catching a space?
	}
	return ret[:len(ret)-1], nil
}

// loadSlackTokentries to return a slack key (from flag or file)
// This function reads from globals (flags)
func loadSlackToken() (slack string, err error) {
	if *flagSlackToken == "" {
		slackFileContents, err := ioutil.ReadFile(*flagSlackTokenFile)
		if err != nil {
			return "", err
		}
		slack = string(slackFileContents)

		// BUG(AJ) I just don't like this
		if strings.HasSuffix(slack, "\r\n") {
			log.Warningf("Slack Token ends in whitespace, eliminating two characters (\\r\\n)...")
			slack = strings.TrimSuffix(slack, "\r\n")
		} else if strings.HasSuffix(slack, "\n") {
			log.Warningf("Slack Token ends in whitespace, eliminating one character (\\n)...")
			slack = strings.TrimSuffix(slack, "\n")
		}
	} else {
		slack = *flagSlackToken
	}
	return slack, nil
}

// loadGitHubToken tries to return a github key (from flag or file)
// This function reads from globals (flags)
func loadGitHubToken() (github string, err error) {
	if *flagGitHubToken == "" {
		githubFileContents, err := ioutil.ReadFile(*flagGitHubTokenFile)
		if err != nil {
			return "", err
		}
		github = string(githubFileContents)
		if strings.HasSuffix(github, "\r\n") {
			log.Warningf("GitHub Token ends in whitespace, eliminating two characters (\\r\\n)...")
			github = strings.TrimSuffix(github, "\r\n")
		} else if strings.HasSuffix(github, "\n") {
			log.Warningf("GitHub Token ends in whitespace, eliminating one character (\\n)...")
			github = strings.TrimSuffix(github, "\n")
		}
	} else {
		github = *flagSlackToken
	}
	return github, nil
}

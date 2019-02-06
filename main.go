// Package main provides a service-like program, issuebot, that listens to a slack channel and allows users to create new issues on a github repo
package main

import (
	"flag"
	"fmt"
	"github.com/shomali11/slacker"
	"os"
	"os/signal"
	"time"
)

// running is a flag that when set to false, tells all goroutines to exit nicely
var running = true

var (
	// flag_org is the name of the org the bot has access to
	flag_org = flag.String("org",
		"",
		"Organization bot has access to")

	// flag_auth provides a file bywhich to load authorized users
	flag_auth = flag.String("auth",
		"./userlist",
		"What file contains a list of authorized users")

	// flag_slack_token will provide a slack token manually
	flag_slack_token = flag.String("slack_token",
		"",
		"Specify the slack token")

	// github_token will provide a github_token manually
	flag_github_token = flag.String("github_token",
		"",
		"Specify the github oauth token")
)

// verifyFlagsSanity just does a basic check on provided flags
func verifyFlagsSanity() {
	if len(*flag_org) == 0 {
		fmt.Printf("You must specify an organizatsion, see --help\n")
		os.Exit(1) // "Panic is for your error, os.Exit is for user error" - Not_a_Golfer
	}
}

// init is called automatically by go
func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: issuebot --org=? [optional flags]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	verifyFlagsSanity()
}

func main() {
	// TODO: check if files are there?
	// TODO: check if token files are there if no tokens?

	/*************** NOTES *****************/
	//BotCommand: Usage() string, Definition() *CommandDefinition, Match(text string) (*proper.Properties, bool), Tokenize() []*commander.Token, Execute(request Request, response ResponseWriter)
	//CommandDefinition: Description, Example (strings), AuthorizationRequired bool, AuthorizedUsers []string, Handler func(request Request, response responseWriter)

	// the responseWriter you get Reply(text string, option ...ReplyOption), ReportError(err error), Typing(), RTM() *slack.RTM, Client() *slack.Client

	// Slacker type is the api bot and handler

	// new client
	// new command
	// register command (example 4/3, 2, 1... parameters, descriptions, etc)
	// context + listen
	// ex. 5 respond w/ error
	// ex. 8 adds a timeout
	// ex. 12 authorization very simple
	// ex. 13 default command

	_ = &slacker.CommandDefinition{}
	/*************** END NOTES AND BS ************/

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	timeNow := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	fmt.Println("Ready")
	for signalRecvd := range signalChannel {
		newTime := time.Now()
		if newTime.Sub(timeNow) < 1000*time.Millisecond {
			fmt.Printf("\nExiting...\n ")
			running = false
			break
		}
		timeNow = newTime
		fmt.Printf("\nReceived a signal: %v\n", signalRecvd)
		fmt.Printf("Reloading auth'ed users\n")
		fmt.Printf("Send again <1 second to exit cleanly\n")
	}
}

// TODO: Refactor this todo list
// TODO: Issue log would be cool on custom response to log channel
// TODO: It would be cool if you could use a particular user's github credentials from slack but that's a part of a custom auth feature
// TODO: Could use oauth but oauth reg on github was requiring a larger-scoped registration process (website, etc)- the super benefit of this is that it restrict users to repos they have access to, which would eliminate some of the problems with github's over-scoped token situation (below). This would also allow us to use the suggest log channel as a way to communicate with issues.
// TODO: the above would change the keyword too
// TODO: unfortunately, github scopes are _not_ granular. for issues, you get +rw on code, pull reqs, wikis, settings, webhooks, deploy keys. this is a 2yr mega thread on github.com/dear-githu[M#Èb

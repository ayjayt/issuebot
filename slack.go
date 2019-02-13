package main

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/ayjayt/slacker"
	"github.com/gravitational/trace"
	"github.com/mailgun/log"
)

const (
	// TimeoutSeconds is how much time Slackbot gives GitHubBot
	TimeoutSeconds = 4
)

var (
	// ErrBadParams is returned when user fails to properly format command arguments
	ErrBadParams = errors.New("user improperly formated command arguments")
)

var (
	parseRegex *regexp.Regexp
)

func init() {
	// See bottom of file to walk-through regex
	parseRegex = regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)" "([^"\\]*(?:\\.[^"\\]*)*)" "([^"\\]*(?:\\.[^"\\]*)*)"`)
}

// paraseParams will collect slack command args as one string and parse it because slacker's built in parser isn't sufficient.
func parseParams(monoParam string) (repo string, title string, body string, err error) {

	resultSlice := parseRegex.FindStringSubmatch(monoParam)
	if (len(resultSlice) != 4) || (len(resultSlice[1]) == 0) || (len(resultSlice[2]) == 0) || (len(resultSlice[3]) == 0) {
		return "", "", "", trace.Wrap(ErrBadParams)
	}
	return resultSlice[1], resultSlice[2], resultSlice[3], nil

}

// TODO: It would be superior if the pkg slacker had
// a) custom auth
// b) a better parser
// c) custom usage

type slackBot struct {
	client      *slacker.Slacker
	gBot        *GitHubIssueBot
	token       string
	authedUsers []string
}

// newSlackBot sets up a connection to slack and registers commands
func newSlackBot(ctx context.Context, token string, authedUsers []string, gBot *GitHubIssueBot) (*slackBot, error) {

	// Making a dynamic "Description" message for our slackbot
	var descriptionString strings.Builder
	descriptionString.WriteString("Creates a new issue on github for ")
	descriptionString.WriteString(gBot.GetOrg())
	descriptionString.WriteString("/YOUR_REPO")

	sBot := &slackBot{
		client:      slacker.NewClient(token),
		gBot:        gBot,
		token:       token,
		authedUsers: authedUsers,
	}

	newIssue := &slacker.CommandDefinition{
		Description: descriptionString.String(),
		Example:     "new \"repo\" \"issue title\" \"issue body\"",
		Handler:     sBot.createNewIssue,
	}

	// Register command
	sBot.client.Command("new <all>", newIssue)

	err := sBot.client.Listen(ctx)
	return sBot, trace.Wrap(err)
}

// createNewIssue is the callback from the user's "New" command on Slack
func (s *slackBot) createNewIssue(r slacker.Request, w slacker.ResponseWriter) {
	// TODO: probably better to be part of an array of CommandDefinitions on slackBot struct then a reciever

	// Sort out parameter
	allParams := r.StringParam("all", "")
	repo, title, body, err := parseParams(allParams)
	if err != nil {
		// This is strictly a user error so log it as info
		log.Infof("Slackbot command user error: %v", err)
		w.ReportError(errors.New("You must specify repo, title, and body for new issue! All in quotes."))
		return
	}

	// Lets try to create a new issue // TODO: ctx with Timeout
	subCtx, cancel := context.WithTimeout(r.Context(), time.Second*TimeoutSeconds)
	defer cancel()

	issue, err := s.gBot.NewIssue(subCtx, repo, title, body)
	if err != nil || subCtx.Err() != nil {
		w.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
		if err != nil {
			log.Infof("Error with gBot.NewIssue: %v", err)
			log.Infof(trace.DebugReport(err))
		} else if subCtx.Err() != nil {
			log.Infof("Error with gBot.NewIssue: %v", err)
			log.Infof(trace.DebugReport(err))
		}
		return
	}
	w.Reply(issue.Url)
	return
}

// Let's build the regex: https://stackoverflow.com/a/6525975
// [^"\\]* <-- Find any non-" and non-\ (tokens) any number of times
// \\. <-- Find any escaped character any number of times
// (?:\\.[^"\\]*)* <-- Find any escape character followed by non-token any number of times... any number of times
// [^"\\]*(?:\\.[^"\\]*)* <-- same as above but it's okay if it's preceeded by non-tokens
// The regex is that repeated 3 times, matched, and surrounded by parenthesis so you can catch, for example:
// "Hello World!" "Backslashes \"\\\" are great" "End":
// Hello World!
// Backslashes "\" are great
// End

package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/ayjayt/slacker"
	"github.com/gravitational/trace"
	"github.com/mailgun/log"
)

const (
	// TimeoutSeconds is how much time Slackbot gives GitHubBot
	TimeoutSeconds = 8
)

var (
	// ErrBadParams is returned when user fails to properly format command arguments
	ErrBadParams = errors.New("user improperly formated command arguments")
)

var (
	// parseRegex will find three quoted strings
	parseRegex *regexp.Regexp
	// escapeRegex will remove one backslash
	escapeRegex *regexp.Regexp
)

func init() {
	// See bottom of file to walk-through regex
	parseRegex = regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)" "([^"\\]*(?:\\.[^"\\]*)*)" "([^"\\]*(?:\\.[^"\\]*)*)"`)
	escapeRegex = regexp.MustCompile(`\\(.)`)
}

// TODO: It would be superior if the pkg slacker had
// a) custom auth
// b) a better parser
// c) custom usage

// BotLink describes a relationships between two persistent API connections.
type BotLink struct {
	sBot *slacker.Slacker // TODO: Maybe a CommandRecevier interface?
	gBot IssuePoster
}

// paraseParams parses an argument string because slacker's parser is limited.
func parseParams(monoParam string) (repo string, title string, body string, err error) {

	resultSlice := parseRegex.FindStringSubmatch(monoParam)
	if (len(resultSlice) != 4) || (len(resultSlice[1]) == 0) || (len(resultSlice[2]) == 0) || (len(resultSlice[3]) == 0) {
		return "", "", "", trace.Wrap(ErrBadParams)
	}

	getMatch := func(matched string) string { return matched }
	deEscape := func(escaped string) string { return escapeRegex.ReplaceAllStringFunc(escaped, getMatch) }

	return deEscape(resultSlice[1]), deEscape(resultSlice[2]), deEscape(resultSlice[3]), nil

}

// createNewIssue is the callback containing logic for "new" command on Slack.
func (s *BotLink) createNewIssue(r slacker.Request, w slacker.ResponseWriter) {

	allParams := r.StringParam("all", "")
	repo, title, body, err := parseParams(allParams)
	if err != nil {
		// This is strictly a user error so log it as info
		log.Infof("Slackbot command user error: %v", err)
		w.ReportError(errors.New("You must specify repo, title, and body for new issue! All in quotes."))
		return
	}

	subCtx, cancel := context.WithTimeout(r.Context(), time.Second*TimeoutSeconds)
	defer cancel()

	issue, err := s.gBot.NewIssue(subCtx, repo, title, body)
	if err != nil || subCtx.Err() != nil {
		if err != nil {
			w.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
			log.Infof("Error with gBot.NewIssue: %v", err)
			log.Infof(trace.DebugReport(err))
		} else if subCtx.Err() != nil {
			w.ReportError(errors.New("Your request timed out"))
			log.Infof("Error with gBot.NewIssue: %v", err)
			log.Infof(trace.DebugReport(err))
		}
		return
	}
	w.Reply(issue.Url)
	return
}

// slackBotHelper defines a BotLink and Slacker (bot) type, and calls Slacker.Listen
func slackBotHelper(ctx context.Context, token string, authedUsers []string, gBot *GitHubIssueBot) (*BotLink, error) {

	botLink := &BotLink{
		sBot: slacker.NewClient(token),
		gBot: gBot,
	}

	newIssue := &slacker.CommandDefinition{
		Description: fmt.Sprintf("Creates a new issue on github for %v/YOUR_REPO", gBot.GetOrg()),
		Example:     "new \"repo\" \"issue title\" \"issue body\"",
		Handler:     botLink.createNewIssue,
	}

	// Register command
	botLink.sBot.Command("new <all>", newIssue)

	err := botLink.sBot.Listen(ctx)
	return botLink, trace.Wrap(err)
}

// Let's build the regex: https://stackoverflow.com/a/6525975
// [^"\\]* <-- Find any non-" and non-\ (tokens) any number of times
// \\. <-- Find any escaped character
// (?:\\.[^"\\]*)* <-- Find any escaped character followed by non-token any number of times... any number of times
// [^"\\]*(?:\\.[^"\\]*)* <-- same as above but it's okay if it's preceeded by non-tokens
// The regex is that repeated 3 times, matched, and surrounded by parenthesis so you can catch, for example:
// "Hello World!" "Backslashes \"\\\" are great" "End":
// Hello World!
// Backslashes "\" are great
// End

package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
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

// SlackBot is a wrapper for the underlying slackbot to include some important variabales
type SlackBot struct {
	sBot    *slacker.Slacker
	gBot    *GitHubIssueBot
	wg      *sync.WaitGroup
	running bool
}

// TODO: It would be superior if the pkg slacker had
// a) custom auth
// b) a better parser
// c) custom usage

// paraseParams parses an argument string into three quoted strings.
// NOTE: The built-in parser only delimits by space, which isn't sufficient.
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
func (s *SlackBot) createNewIssue(r slacker.Request, w slacker.ResponseWriter) {
	if s.CheckRun(w) {
		defer s.Done()
	} else {
		return
	}
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
		}
		if subCtx.Err() != nil {
			w.ReportError(errors.New("Your request timed out"))
			log.Infof("Error with gBot.NewIssue: %v", subCtx.Err())
			log.Infof(trace.DebugReport(subCtx.Err()))
		}
		return
	}
	w.Reply(issue.Url)
	return
}

// newSlackBot BotLink and Slacker (bot) type, and calls Slacker.Listen
func newSlackBot(token string, authedUsers []string, gBot *GitHubIssueBot) *SlackBot {

	slackBot := &SlackBot{
		sBot:    slacker.NewClient(token),
		gBot:    gBot,
		wg:      &sync.WaitGroup{},
		running: false,
	}

	newIssue := &slacker.CommandDefinition{
		Description:           fmt.Sprintf("Creates a new issue on github for %v/YOUR_REPO", gBot.GetOrg()),
		Example:               `new "repo" "issue title" "issue body"`,
		AuthorizationRequired: true,
		AuthorizedUsers:       authedUsers,
		Handler:               slackBot.createNewIssue,
	}

	// Register command
	slackBot.sBot.Command("new <all>", newIssue)

	return slackBot
}

// Listen calls Listen on the underlying slackbot
func (s *SlackBot) Listen(ctx context.Context) error {
	err := s.sBot.Listen(ctx)
	s.running = true
	return trace.Wrap(err)
}

// EmptyQueue is called to stop slack commands from starting and locking the program into running.
// Once all commands are done, the waitgroup can be passed.
func (s *SlackBot) EmptyQueue() {
	s.running = false
	s.wg.Wait() // TODO: Add a timeout maybe?
}

// CheckRun will check to see if you should be using the waitgroup.
func (s *SlackBot) CheckRun(w slacker.ResponseWriter) bool {

	if !s.running { // Don't lock if we're not running
		w.ReportError(errors.New("I'm shutting down"))
		return false
	}
	s.wg.Add(1)
	if !s.running { // Unlock if we canceled between the first if statement and now
		w.wg.Done()
		w.ReportError(errors.New("I'm shutting down"))
		return false
	}
	return true
}

// Done is just a wrapper for sync.WaitGroup.Done
func (s *SlackBot) Done() {
	s.wg.Done()
}

// Let's build the regex: https://stackoverflow.com/a/6525975
// [^"\\]* <-- Find any non-" and non-\ (tokens) any number of times
// \\. <-- Find any escaped character
// (?:\\.[^"\\]*)* <-- Find any escaped character followed by non-token any number of times... any number of times
// [^"\\]*(?:\\.[^"\\]*)* <-- same as above but it's okay if it's preceeded by non-tokens
// The regex is that repeated 3 times, matched, and surrounded by quotations so you can catch, for example:
// "Hello World!" "Backslashes \"\\\" are great" "End":
// Hello World!
// Backslashes "\" are great
// End

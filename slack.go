package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"sync"
	"time"

	"github.com/ayjayt/slacker"
	"github.com/gravitational/trace"
	"github.com/mailgun/log"
	"github.com/shomali11/proper"
)

// BUG(AJ) EXPLICITLY CATCH WHEN IT DOESN'T WORK - when what doesn't work? there's bugs... the main bug is that... if register is first command it works

const (
	// TimeoutSeconds is how much time Slackbot gives GitHubBot
	TimeoutSeconds = 8
)

var (
	// ErrBadParams is returned when user fails to properly format command arguments
	ErrBadParams = errors.New("user improperly formated command arguments")
)

const (
	userTokenFile = "./usertokens"
)

var (
	// parseRegex will find three quoted strings
	issueRegex *regexp.Regexp
	// escapeRegex will remove one backslash
	escapeRegex *regexp.Regexp
)

func init() {
	// See bottom of file to walk-through (partial)
	issueRegex = regexp.MustCompile(`^\s*(?:<@(\S+)>)?\s*new\s+"([^"\\]*(?:\\.[^"\\]*)*)"\s+"([^"\\]*(?:\\.[^"\\]*)*)"\s+"([^"\\]*(?:\\.[^"\\]*)*)"\s*$`)
	escapeRegex = regexp.MustCompile(`\\(.)`)
}

// SlackBot is a wrapper for the underlying slackbot to include some important variabales
type SlackBot struct {
	sBot                *slacker.Slacker
	gBots               sync.Map          // TODO: By user, maybe a slack user type
	userTokens          map[string]string // TODO: Protect this against concurrent access
	userTokensFileLock  sync.Mutex
	userTokensFileQueue int
	// TODO: default gBot based on token
	wg      *sync.WaitGroup
	running bool
	botID   string
}

// readStore finds the file storing tokens and demarshals it into gBots sync.Map
// TODO: make the a seperate type (any interface with concurrent read write I guess) so it can be easily changed out
func (s *SlackBot) readStore() error { // should be streaming
	s.userTokens = make(map[string]string)
	userTokenFileContents, err := ioutil.ReadFile(userTokenFile)
	if err != nil {
		return trace.Wrap(err)
	}
	err = json.Unmarshal(userTokenFileContents, &s.userTokens)
	return trace.Wrap(err)
}

// writeStore marshals tokens from gBots sync.Map into a file
// TODO: make the a seperate type (any interface with concurrent read write I guess) so it can be easily changed out
func (s *SlackBot) writeStore() {
	if s.userTokensFileQueue > 2 {
		// No need to write that many times
		return
	}
	go func() {
		s.userTokensFileQueue++
		s.userTokensFileLock.Lock()
		userTokenFileContents, err := json.Marshal(s.userTokens)
		if err != nil {
			log.Errorf(trace.DebugReport(err))
			return
		}
		err = ioutil.WriteFile(userTokenFile, userTokenFileContents, 0600)
		if err != nil {
			log.Errorf(trace.DebugReport(err))
			return
		}
		s.userTokensFileLock.Unlock()
		s.userTokensFileQueue--
	}()
}

// newIssueParser takes a whole command and matches and creates three params. It's a custom parser for one command. TODO: add snippets and default detection
func (s *SlackBot) newIssueParser(text string) (*proper.Properties, bool) {
	log.Infof("in newIssueParser for %v with %v", s.botID, text)
	resultSlice := issueRegex.FindStringSubmatch(text) // TODO remove all botnames that aren't quoted before this

	var wordOffset = 0
	if (len(resultSlice) < 2) || (len(resultSlice[1]) == 0) || (len(resultSlice[2]) == 0) || (len(resultSlice[3]) == 0) {
		return nil, false
	}
	if len(resultSlice) == 5 {
		if (len(resultSlice[4]) == 0) || (resultSlice[1] != s.botID) {
			return nil, false
		} else {
			wordOffset = 1
		}
	}
	getMatch := func(matched string) string { return matched }
	deEscape := func(escaped string) string { return escapeRegex.ReplaceAllStringFunc(escaped, getMatch) }
	parameters := make(map[string]string)
	parameters["repo"] = deEscape(resultSlice[1+wordOffset])
	parameters["title"] = deEscape(resultSlice[2+wordOffset])
	parameters["body"] = deEscape(resultSlice[3+wordOffset])
	return proper.NewProperties(parameters), true

}

/*************
* The following are for routing slack users to their github client with github client token CRUD functions
*************/

// GetGBot can find the relevant github client for a particular slack user. or initialize it
func (s *SlackBot) GetGBot(r slacker.Request) *GitHubIssueBot {
	log.Infof("Getting bot for : %v", r.Event().User)
	ret, ok := s.gBots.Load(r.Event().User)
	if !ok {
		diskCheck, ok := s.userTokens[r.Event().User] // TODO: write/read concurrency issues. Only one at a time, or mutexes, or a queue.
		if ok {
			ret = NewGitHubIssueBot(r.Context(), diskCheck)
			// TODO: this needs testing
			s.gBots.Store(r.Event().User, ret)
		} else {
			return nil
		}
	}
	return ret.(*GitHubIssueBot)
}

// SetGBot will create a new user-github association
func (s *SlackBot) SetGBot(r slacker.Request, gBot *GitHubIssueBot) bool {
	// TODO: maybe testing should be in here -- definitely
	_, ok := s.gBots.Load(r.Event().User)
	if ok {
		return !ok
	}
	s.gBots.Store(r.Event().User, gBot)
	s.userTokens[r.Event().User] = gBot.token
	s.writeStore()
	return true
}

// DeleteGBot will create a new user-github association
func (s *SlackBot) DeleteGBot(r slacker.Request) {
	log.Infof("Deleting bot") // TODO: all log ettiquette
	s.gBots.Delete(r.Event().User)
	delete(s.userTokens, r.Event().User)
	s.writeStore() // TODO: somehow I thought a read-writer would be good for this.
}

/*************
* The following are command definitions
*************/

// createNewIssue is the callback containing logic for "new" command on Slack.
func (s *SlackBot) createNewIssue(r slacker.Request, w slacker.ResponseWriter) {
	if s.CheckRun(w) {
		defer s.Done()
	} else {
		return
	}
	repo := r.StringParam("repo", "")
	title := r.StringParam("title", "")
	body := r.StringParam("body", "")

	subCtx, cancel := context.WithTimeout(r.Context(), time.Second*TimeoutSeconds)
	defer cancel()

	client := s.GetGBot(r)
	if !s.CheckClient(w, client) { // TODO: This could be in auth
		return
	}
	issue, err := client.NewIssue(subCtx, repo, title, body) // TODO: you'll panic if they delete while doing this
	if err != nil || subCtx.Err() != nil {
		if err != nil {
			w.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
			log.Infof("new issue error: %v", err)
			log.Infof(trace.DebugReport(err))
		}
		if subCtx.Err() != nil {
			w.ReportError(errors.New("Your request timed out"))
			log.Infof("new issue error: %v", subCtx.Err())
			log.Infof(trace.DebugReport(subCtx.Err()))
		}
		return
	}
	w.Reply(issue.Url)
	return
}

func (s *SlackBot) registerUser(r slacker.Request, w slacker.ResponseWriter) {
	// BUG(AJ) THIS WILL REPEAT IF YOU DO IT RIGHT AWAY OR SOMETHING EVNE IF THEY FIND YOU
	if s.CheckRun(w) {
		defer s.Done()
	} else {
		return
	}
	token := r.StringParam("token", "")
	if token == "" {
		w.ReportError(errors.New("You must specify a token"))
		return
	}
	gBot := NewGitHubIssueBot(r.Context(), token) // TODO: is this context... the global context?
	subCtx, cancel := context.WithTimeout(r.Context(), time.Second*TimeoutSeconds)
	defer cancel()

	name, login, err := gBot.CheckToken(subCtx)
	if err != nil {
		w.ReportError(errors.New("Token didn't work"))
		return
	}
	ok := s.SetGBot(r, gBot)
	if !ok {
		w.ReportError(errors.New("User already registered, please delete first."))
		return
	}
	w.Reply(fmt.Sprintf("User successfully registered: %v, %v", name, login))
	return
}

func (s *SlackBot) deleteUser(r slacker.Request, w slacker.ResponseWriter) {
	if s.CheckRun(w) {
		defer s.Done()
	} else {
		return
	}
	s.DeleteGBot(r)
	w.Reply("If you had registered, you are no longer.")
	return
}

/*************
* The following are initializers
*************/

// newSlackBot BotLink and Slacker (bot) type, and calls Slacker.Listen
func newSlackBot(token string, authedUsers []string) *SlackBot {

	slackBot := &SlackBot{
		sBot:    slacker.NewClient(token),
		wg:      &sync.WaitGroup{},
		running: false,
	}
	err := slackBot.readStore()
	if err != nil {
		log.Errorf(trace.DebugReport(err))
	}
	newIssue := &slacker.CommandDefinition{
		Description:           "Creates a new issue on github for repo specified",
		Example:               `new "repo" "issue title" "issue body"`,
		AuthorizationRequired: false,
		CustomParser:          slackBot.newIssueParser,
		Handler:               slackBot.createNewIssue,
	}

	registerUser := &slacker.CommandDefinition{
		Description:           "Associate a github token with a user",
		AuthorizationRequired: false,
		Handler:               slackBot.registerUser,
	}

	deleteUser := &slacker.CommandDefinition{
		Description:           "Disassociate a github token with a user",
		AuthorizationRequired: false,
		Handler:               slackBot.deleteUser,
	}

	// Register command
	slackBot.sBot.Command("register <token>", registerUser)
	slackBot.sBot.Command("unregister", deleteUser)
	slackBot.sBot.Command("new <repo> <title> <body>", newIssue)
	slackBot.sBot.Init(func(s *SlackBot) func() {
		return func() {
			// TODO: add context
			res, err := s.sBot.Client().AuthTest()
			if err != nil {
				trace.Wrap(err)
				log.Errorf("Error trying to get botname: %v", trace.DebugReport(err))
			}
			s.botID = res.UserID
		}
	}(slackBot))
	return slackBot
}

// Listen calls Listen on the underlying slackbot
func (s *SlackBot) Listen(ctx context.Context) error {
	s.running = true
	err := s.sBot.Listen(ctx)
	return trace.Wrap(err)
}

/*************
* The following are helper functions and often write directly to slack
*************/

func (s *SlackBot) CheckClient(w slacker.ResponseWriter, client *GitHubIssueBot) bool {
	if client != nil {
		return true
	}
	w.ReportError(errors.New("You must register first, see `help` command"))
	return false
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
		s.wg.Done()
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

package main

import (
	"context"
	"errors"
	"strings"

	"github.com/ayjayt/slacker"
	"github.com/gravitational/trace"
	"github.com/mailgun/log"
)

// parseParam parses the single parameter into 3 because the built-in command parser can't handle whitespace. parseParam just wants to see three phrases in quotations.
func parseParam(allParam string) (repo string, title string, body string, ok bool) {

	// threeParams is the three phrases
	var threeParams [3]string
	// paramCount tracks how many phrases we've found
	var paramCount int = 0
	// quoteSwitch tracks if we're inside or outside quotations
	var quoteSwitch bool = false
	// escapeSwitch knows if we should process the next character specially or not
	var escapeSwitch bool = false

	// start is the index where last phrase started
	var start int = 0
	for i, c := range allParam {
		if (i == 0) && (c != '"') { // first character has to be a quote
			return "", "", "", false
		} else if i == 0 { // first character AND it was a quote
			quoteSwitch = true
			start = i + 1
		} else if (!escapeSwitch) && (c == '\\') { // we'll escape the next character if we weren't escaped
			escapeSwitch = true
		} else if escapeSwitch { // turn off escapeSwitch and move on (we have escaped the current character)
			escapeSwitch = false
		} else if c == '"' { // we've encountered a non-escaped quote
			if quoteSwitch { // we were in quotes, now we're out
				threeParams[paramCount] = allParam[start:i]
				paramCount += 1
			} else { // we are just starting quotes
				start = i + 1
			}
			quoteSwitch = !quoteSwitch
		}
	}

	if paramCount != 3 {
		return "", "", "", false
	}
	return threeParams[0], threeParams[1], threeParams[2], true
}

// openBot just starts the bot with the callback.
func openBot(ctx context.Context, token string, authedUsers []string, gBot *GitHubIssueBot) (err error) {

	// Making a dynamic "Description" message for our slackbot
	var descriptionString strings.Builder
	descriptionString.WriteString("Creates a new issue on github for ")
	descriptionString.WriteString(gBot.GetOrg())
	descriptionString.WriteString("/YOUR_REPO")

	sBot := slacker.NewClient(token)

	// newCommand is built by a callback factory to attach the CB to a certain GitHubIssueBot
	newCommand := func(gBot *GitHubIssueBot) func(slacker.Request, slacker.ResponseWriter) {
		return func(request slacker.Request, response slacker.ResponseWriter) {

			// Note: This supports multiple commands but not "", and I didn't want to override/reimplement the interfaces due to time-cost so I implemented a monolothic parameter and parse it myself.
			allParam := request.StringParam("all", "")
			repo, title, body, ok := parseParam(allParam)

			// Params were bad
			if !ok {
				response.ReportError(errors.New("You must specify repo, title, and body for new issue! All in quotes."))
				return
			}

			// Lets try to create a new issue
			var URL string
			URL, err = gBot.NewIssue(ctx, repo, title, body) // TODO: issue structure instead of three parameters
			if err != nil {
				response.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
				return
			}
			// Issue was good
			response.Reply(URL)
			return
		}
	}(gBot) // call factory function with parameters bassed to openBot

	newIssue := &slacker.CommandDefinition{
		Description: descriptionString.String(),
		Example:     "new \"repo\" \"issue title\" \"issue body\"",
		Handler:     newCommand,
	}

	// Reigster command
	sBot.Command("new <all>", newIssue)

	log.Infof("Starting slack bot listen...")
	err = sBot.Listen(ctx)
	log.Infof("bot.Listen(ctx) returned")

	return trace.Wrap(err)
}

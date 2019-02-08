package main

import (
	"context"
	"errors"
	"github.com/ayjayt/slacker"
	"github.com/mailgun/log"
	"strings"
	"sync"
)

// processParam processes command parameters manualy because the built-in command processor is extremely weak
func processParam(allParam string) (repo string, title string, body string, ok bool) {
	// look for three sets of quotes
	// loop through allParam, keeping track of open and closing quotes, and allowing us skip processing one character at a time (escape)
	var threeParams [3]string
	var paramCount int = 0
	var quoteSwitch bool = false
	var escapeSwitch bool = false

	// start is where the last quote started
	var start int = 0
	log.Debugf("allParam: %v", allParam)
	for i, c := range allParam {
		log.Debugf("i, c: %d, %c", i, c)
		if (i == 0) && (c != '"') { // first character has to be a quote
			log.Debugf("Bad start to allParam")
			return "", "", "", false
		} else if i == 0 { // first character was a quote
			quoteSwitch = true
			start = i + 1
			log.Debugf("Open the quotes!")
		} else if (!escapeSwitch) && (c == '\\') { // we'll escape the next character if we weren't escaped
			escapeSwitch = true
			log.Debugf("Next character literal")
		} else if escapeSwitch { // turn off escapeSwitch and move on (we have escaped the current character)
			escapeSwitch = false
			log.Debugf("Character was literal")
		} else if c == '"' { // we've encountered a non-escaped quote
			if quoteSwitch { // we were in quotes, now we're out
				threeParams[paramCount] = allParam[start:i]
				paramCount += 1
				log.Debugf("Turn off quotes! %v", threeParams[paramCount-1])
			} else { // we are just starting quotes
				start = i + 1
				log.Debugf("Turn on quotes!")
			}
			quoteSwitch = !quoteSwitch
		}
	}

	if paramCount != 3 {
		return "", "", "", false
	}
	return threeParams[0], threeParams[1], threeParams[2], true
}

// OpenBot just starts the bot with the callback. BUG(AJ) Warning- this bot library doesn't like concurrency. This library is written like we're in node.js.
func openBot(ctx context.Context, token string, authedUsers []string, waitForCb sync.WaitGroup, gBot *GitHubIssueBot) (err error) {

	// Making a dynamic "Description" message for our slackbot
	var descriptionString strings.Builder
	descriptionString.WriteString("Creates a new issue on github for ")
	descriptionString.WriteString(gBot.GetOrg())
	descriptionString.WriteString("/YOUR_REPO")

	// At least it uses a client... TODO: investigate if you can use this to control it concurrently
	sBot := slacker.NewClient(token)

	// newCommand is built by a callback factory to attach the CB to a certain waitgroup and GitHubIssueBot
	newCommand := func(waitForCb sync.WaitGroup, gBot *GitHubIssueBot) func(slacker.Request, slacker.ResponseWriter) {
		return func(request slacker.Request, response slacker.ResponseWriter) {

			// running has let us know to stop taking new issues for the moment
			if !running { // TODO: good candidate for context
				response.ReportError(errors.New("Issuebot is starting up or shutting down, try again in a few seconds."))
				return
			}

			// Okay, we're going to start network ops, so please wait until we're done
			waitForCb.Add(1)
			defer waitForCb.Done()

			// Note: This supports multiple commands but not "", and I didn't want to override/reimplement the interfaces due to time-cost so I implemented a monolothic parameter and parse it myself.
			allParam := request.StringParam("all", "")
			repo, title, body, ok := processParam(allParam)

			// Params were bad
			if !ok {
				response.ReportError(errors.New("You must specify repo, title, and body for new issue! All in quotes."))
				return
			}

			// Lets try to create a new issue
			var URL string
			URL, err = gBot.NewIssue(repo, title, body)
			if err != nil {
				response.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
				return
			}
			// Issue was good
			response.Reply(URL)
			return
		}
	}(waitForCb, gBot) // call factory function with parameters bassed to openBot

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

	return err
}

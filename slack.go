package main

import (
	"github.com/shomali11/slacker"
)

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

func init() {
	_ = &slacker.CommandDefinition{}
	// start with example give, and move on
}

/*************** END NOTES AND BS ************/

// Command issuebot provides a service-like program, issuebot, that listens to a slack channel and allows users to create new issues on a github repo
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
)

// init() prepares a console logger
func init() {
	// Note: You can load more loggers after this Init.
	console, _ := log.NewLogger(log.Config{"console", "debug"}) // note: debug, info, warning, error
	log.Init(console)
}

// run contains the main program logic
func run(ctx context.Context) error {

	cfg, err := flagHelper() // flags.go
	if err != nil {
		return err
	}

	gitHubBot := NewGitHubIssueBot(ctx, cfg.gitHubToken) // github.go
	if err := gitHubBot.CheckOrg(ctx, cfg.org); err != nil {
		return err
	}

	slackBotErr := make(chan error)
	go func() {
		// NOTE: No need for returned botLink, but I'm sure it will be useful
		if _, err := slackBotHelper(ctx, cfg.slackToken, cfg.authedUsers, gitHubBot); err != nil {
			if err != context.Canceled {
				slackBotErr <- trace.Wrap(err)
			}
			close(slackBotErr)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)

	select {
	case <-signalChannel:
		log.Infof("Received interrupt signal")
		// TODO: Healthy exit
	case <-ctx.Done():
		// NOTE: context.CancelFunc is a hard kill
	case err := <-slackBotErr:
		return err
	}
	return nil
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		log.Errorf("Program couldn't start: %v", err)
		log.Errorf("Trace.Debug(err): %v", trace.DebugReport(err))
		os.Exit(1)
	}

}

// TODO: Refactor this "future" todo list
// TODO: Issue log would be cool on custom response to log channel
// TODO: It would be cool if you could use a particular user's github credentials from slack but that's a part of a custom auth feature
// TODO: Could use oauth but oauth reg on github was requiring a larger-scoped registration process (website, etc)- the super benefit of this is that it restrict users to repos they have access to, which would eliminate some of the problems with github's over-scoped token situation (below). This would also allow us to use the suggest log channel as a way to communicate with issues.
// TODO: the above would change the keyword too
// TODO: unfortunately, github scopes are _not_ granular. for issues, you get +rw on code, pull reqs, wikis, settings, webhooks, deploy keys. this is a 2yr mega thread on github.com/dear-githu[M#Èb
// TODO: one way to turn this into an interface would be to create a new bot type that could be initialized with a slack org(s) and github org(s) but I feel like it would just be better to run seperate processes -- although multiple slack orgs and github orgs would be good (although multiple github orgs if users supply their own keys too)
// TODO: needs timeout on network ops
// TODO: there needs to be better than using a global to send signals

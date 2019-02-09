// Command issuebot provides a service-like program, issuebot, that listens to a slack channel and allows users to create new issues on a github repo
package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gravitational/trace"
	"github.com/mailgun/log"
)

// Var running is a flag that when set to false, tells all goroutines to exit nicely- and new callbacks not to start important network ops
// Maybe a good candidate for context
var running = true

// Func init() prepares a console logger
func init() {
	// Note: You can load more loggers after this Init.
	console, _ := log.NewLogger(log.Config{"console", "debug"}) // note: debug, info, warning, error
	log.Init(console)
}

func main() {
	// WaitGroup so that callbacks can finish before run exits when told to
	var waitForCb sync.WaitGroup
	var err error

	cfg, err := checkAndSetDefaults()
	if err != nil { // flags.go
		log.Errorf("Program couldn't start: %v", err)
		os.Exit(1)
	}

	// run will wait for a signal (SIGINTish), wait for the slackbot to clean up (WaitGroup), and then os.Exit(0)
	// TODO: change to SIGHUP
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	githubBot := NewGitHubIssueBot(cfg.githubToken) // github.go
	githubBot.Connect(ctx)
	if ok, err := githubBot.CheckOrg(ctx, cfg.org); !ok || err != nil {
		if err != nil {
			log.Errorf("Error checking for organization: %v", trace.Wrap(err))
		}
		log.Errorf("Couldn't load or find the org supplied")
		os.Exit(1)
	}

	go func() {
		err = openBot(ctx, cfg.slackToken, cfg.authedUsers, waitForCb, githubBot) // TODO: I really wanna put the WaitGroup+Running in the Context
		if err != nil {
			log.Errorf("Some problem starting the Slack bot: %T: %v", err, err)
			os.Exit(1)
		}
		log.Infof("IssueBot connected for org %v", cfg.org)
	}()

	run(ctx)

	// Don't exit until done- TODO this should be timed-out
	waitForCb.Wait()
	cancel()
}

// run waits for signals to exit or reload information
func run(ctx context.Context) {

	// This is all for catching signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	var timeNow time.Time
	log.Infof("Waitng for signals...")
	// Now we're going to wait on signals from terminal
	for {
		select {
		case signalRecvd := <-signalChannel:
			newTime := time.Now()
			if newTime.Sub(timeNow) < 1000*time.Millisecond {
				log.Infof("Exiting...")
				running = false
				return
			}
			timeNow = newTime
			log.Infof("Received a signal: %v", signalRecvd)
			log.Infof("Reloading auth'ed users and keys")
			// TODO: reload authed users and keys- don't panic on error
			// have they changed? no- don't do it
			// are they there? no- warn
			// else, redo, restart
			log.Infof("Send again <1 second to exit cleanly")
		case <-ctx.Done():
			break
		}
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

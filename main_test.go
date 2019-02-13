package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	. "gopkg.in/check.v1" // not sure why we went this route (with .) but I'm taking hints from gravitational/hello
)

// Check package hooks into normal testing
func TestIssueBot(t *testing.T) { TestingT(t) }

// Check package needs a type
type MainSuite struct{}

var _ = Suite(&MainSuite{})

func (s *MainSuite) SetUpSuite(c *C) {
	// This may be used if test coverage improves
}

func (s *MainSuite) SetUpTest(c *C) {
	running = true
}

func (s *MainSuite) TearDownTest(C *C) {
	running = true
}

// TestRun looks to make sure that signals are working properly
// It's a unique test in that it's not a "fn(param...) = return" type testing structure
func (s *MainSuite) TestRun(c *C) {
	// TODO: race tests
	if runtime.GOOS == "windows" { // TODO: Check, really- see BUG below
		fmt.Println("You are running on windows. There is a bug that may or may not exist preventing it from running correctly. Please either delete this if-block if the test works, or uncomment the skip below if the test does not and submit a pull request or raise an issue.")
		// c.Skip("Test doesn't work on windows")
	}

	// Function for custom comments based on stage
	comment := func(stage string) CommentInterface {
		return Commentf("Run test showed running to be %v when it should be %v %v.", running, !running, stage)
	}

	// Return my PID
	me, err := os.FindProcess(os.Getpid())
	if err != nil {
		c.Errorf("Test itself is failing: %v", err)
		return
	}

	// Marking the stage
	c.Assert(running, Equals, true, comment("before running"))

	// Set a timeout
	timeout := time.NewTimer(3 * time.Second)
	blockChan := make(chan int)
	// WaitGroup could probably be used instead of blockChan somehow
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		blockChan <- 1
		run(ctx)
		close(blockChan) // close channel to communicate to goroutine that run finished
	}()

	<-blockChan                        // Just making sure that goroutine started
	time.Sleep(200 * time.Millisecond) // Effectively a timeout for the run() goroutine to do it's thing

	// BUG(AJ): This os.Process.Signal(os.Interrupt) wont work on windows? https://github.com/golang/go/issues/6720
	// It seems that only the test is affected. ^C is interpreted correctly as os.Interrupt when _listening_, but sending os.Interrupt does != ^C on windows

	// First interrupt- reload
	me.Signal(os.Interrupt)
	time.Sleep(200 * time.Millisecond) // Signal takes time. Think of 200 ms like a timeout.
	c.Assert(running, Equals, true, comment("after one signal"))

	// Second interrupt- exit run()
	me.Signal(os.Interrupt)
	time.Sleep(200 * time.Millisecond) // Signal takes time. Think of 200 ms like a timeout.
	c.Assert(running, Equals, false, comment("after a second signal"))

	// Timeout or finish?
	select {
	case <-blockChan: // this will be closed if run unblocks
		if timeout.Stop() { // returns true if succesful
			return
		}
	case <-timeout.C: // this will be "actuated" if the timeout expires
		c.Error("Test Failure: Timeout. run() never seems to have returned")
		c.FailNow()
		return
	}

}

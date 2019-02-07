package main

import (
	"fmt"
	. "gopkg.in/check.v1" // not sure why we went this route (with .) but I'm taking hints from gravitational/hello
	"os"
	"runtime"
	"testing"
	"time"
)

// Check package hooks into normal testing
func TestIssueBot(t *testing.T) { TestingT(t) }

// Check package needs a type
type MainSuite struct{}

var _ = Suite(&MainSuite{})

func (s *MainSuite) SetUpSuite(c *C) {
	// If you want
}

func (s *MainSuite) SetUpTest(c *C) {
	running = true
}

func (s *MainSuite) TearDownTest(C *C) {
	running = true
}

// TestProcessFlow just looks to make sure that signals are working properly
// It's a unique test in that it's not a fn(param...) = return type structure
func (s *MainSuite) TestRun(c *C) {

	if runtime.GOOS == "windows" { // TODO: Check, really- see BUG below
		fmt.Println("You are running on windows. There is a bug that may or may not exist preventing it from running correctly. Please either delete this if block if the test works, or uncomment the line below this print statement if the test does not and submit a pull request or raise an issue.")
		// c.Skip("Test doesn't work on windows")
	}

	// Because we can't loop through a table
	comment := func(stage string) CommentInterface {
		return Commentf("Run test showed running to be %v when it should be %v %v.", running, !running, stage) // just figured that these variables would be evaluated once when Commentf was called
	}

	me, err := os.FindProcess(os.Getpid())
	if err != nil {
		c.Errorf("Test itself is failing: %v", err)
		return
	}

	// Marking the stage
	c.Assert(running, Equals, true, comment("before running"))

	timeout := time.NewTimer(3 * time.Second)
	blockChan := make(chan int)
	go func() {
		blockChan <- 1
		run()
		close(blockChan) // close channel to communicate to goroutine that run finished
	}()

	<-blockChan                        // Just making sure that coroutine started
	time.Sleep(200 * time.Millisecond) // Effectively a timeout for the run() coroutine to do it's thing

	// BUG(AJ): This os.Process.Signal(os.Interrupt) wont work on windows? https://github.com/golang/go/issues/6720
	// It seems that only the test is affected. ^C is interpreted correctly as os.Interrupt when _listening_, but sending os.Interrupt does != ^C on windows

	me.Signal(os.Interrupt)
	time.Sleep(200 * time.Millisecond) // Signal takes time. Think of 200 ms like a timeout.
	c.Assert(running, Equals, true, comment("after one signal"))

	me.Signal(os.Interrupt)
	time.Sleep(200 * time.Millisecond) // Signal takes time. Think of 200 ms like a timeout.
	c.Assert(running, Equals, false, comment("after a second signal"))

	// Timeout or finish?
	select {
	case <-blockChan: // this will be closed if run unblocks
		if timeout.Stop() { // returns true if succesful stops a timer, I guess
			return
		}
	case <-timeout.C: // this will be "actuated" if the timeout expires
		c.Error("Test Failure: Timeout. run() never seems to have returned")
		c.FailNow()
		return
	}

}

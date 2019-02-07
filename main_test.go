package main

import (
	"os"
	"testing"
	//"fmt"
	. "gopkg.in/check.v1" // not sure why we went this route (with .) but I'm taking hints from gravitational/hello
)

func init() {
	// Default with these off...
	os.Setenv("ISSUEBOT_FLAGS", "NO")
	os.Setenv("ISSUEBOT_LOGS", "NO")
}

// Check package hooks into normal testing
func TestIssuebot(t *testing.T) { TestingT(t) }

// Check package needs a type
type IssueBotSuite struct{}

var _ = Suite(&IssueBotSuite{})

func (s *IssueBotSuite) SetUpSuite(c *C) {
	// If you want
}

func (s *IssueBotSuite) SetUpTest(c *C) {
	// If you want
}

func (s *IssueBotSuite) TearDownTest(C *C) {
	// If you want
}

// gravtional/hello had two seperate tests for the Okay input/out and the Erroneus input/out- I combined it into one for no great reason other than to see if I could.
func (s *IssueBotSuite) TestTableOk(c *C) {
	// struct of name, parameters and expected outputs
	// use Commentf, c.Assert
	// run the function for the whole table
	var err error = nil
	c.Assert(err, IsNil)
}

// No need to do benchmark for this

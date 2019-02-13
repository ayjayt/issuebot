package main

import (
	"testing"

	. "gopkg.in/check.v1" // not sure why we went this route (with .) but I'm taking hints from gravitational/hello
)

// Check package hooks into normal testing
func TestIssueBot(t *testing.T) { TestingT(t) }

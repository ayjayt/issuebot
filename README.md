# issuebot

**issuebot** is a slackbot that allows creating new github issues with a command

## Building

_Hint: `git clone` this repo to your `$GOPATH/src` directory; `go help gopath` and `go help importpath` for more info_

Make commands (there may be more in _Makefile_): 

`make build` 
Outputs binary to _./build/_. This is the default make target.

`make tidy`  
Runs `gofmt` and `govet` on files specified in Makefile's _FILE_ variable

`make test`  
Runs `go test` as expected

`make clean`  
Clears _./build/_ directory


### Dependencies

This is a list of 3rd party dependencies, which you can load with `go get` or migrate manually to your organization's dependency management system.

#### Run

[github.com/shomali11/slacker](https://github.com/shomali11/slacker)  
A simple library that allows you to create slackbots by registering a command string and a callback

[github.com/shurcooL/githubv4](https://github.com/shomali11/slacker)  
An extensive library that creates a go interface for the GitHub APIv4 (GraphQL)

[golang.org/x/oauth2](https://godoc.org/golang.org/x/oauth2)  
The oauth library used by GitHub APIv4 for auth

[github.com/mailgun/log](https://github.com/mailgun/log)  
The logger used by this programmer

#### Test

[gopkg.in/check.v1](https://gopkg.in/check.v1)  
A testing suite to enhance Go's native "testing" module

[github.com/ayjayt/crashtest](https://github.com/ayjayt/crashtest)
A helper library to catch functions that panic or exit and test for error codes

## After Building: Configure and Run

* Create a slack token (refer to slack docs)
* Create a github token (refer to github docs)  
_NB: as of Feb 2019, this can be the personal or the "oauth"- effectively the same (oath), but "oath" registers your "app"_
* Create two files with just the tokens (or override these w/ one of two flags-- see flags with `issuebot --help`):
  * slack_token
  * github_token
* Create a file and list slack users (**BY WHAT**), each on a new line, who can use the issuebot. The default file name is *./userlist* but can be overrided by a flag.
* Run the program (see below for typical use or use `issuebot --help` to see all flags) in _./build/_ and it will do it's best to connect

The benefit of using files for keys and users is that the files can be reloaded by with <kbd>^C</kbd>. Hit <kbd>^C</kbd> twice to exit cleanly.

## Using with Slack

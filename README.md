# issuebot

**issuebot** is a slackbot that allows creating new github issues with a command

# Building

_Hint: `git clone` this repo to your `$GOPATH/src` directory; `go help gopath` and `go help importpath` for more info`_

Make commands (may be more in _Makefile_): 

`make build (default)`  
Outputs binary to _./build/_

`make tidy`  
Runs `gofmt` and `govet` on files specified in Makefile's _FILE_ variable

`make test`  
Runs `go test` as expected

`make clean`  
Clears _./build/_ directory


## Dependencies

This is a list of 3rd party dependencies, which you can load with `go get` or migrate to your organization's dep management system.

[github.com/shomali11/slacker](https://github.com/shomali11/slacker)  
A simple library that allows you to create slackbots by registering a command string and a callback

[github.com/shurcooL/githubv4](https://github.com/shomali11/slacker)  
An extensive library that creates a go interfaces for the GitHub APIv4 (graphQL)

[golang.org/x/oauth2](https://godoc.org/golang.org/x/oauth2)  
The oauth library used by githubv4 for auth

# After Building: Configure and Run

* Create a slack token (refer to slack docs)
* Create a github token (refer to github docs)  
_NB: as of Feb 2019, this can be the personal or the "oauth"- effectively the same (oath), but "oath" registers your "app"_
* Create two files with just the tokens (or override w/ flag-- see below):
  * slack_token
  * github_token
* Create a file and list slack users (**BY WHAT**), each on a new line, who can use the issuebot. This will be specified on the command line.
* Run the program (see below for normal flags or use --help)  in ./build and it will do it's best to connect

# Using with Slack

# Preliminary Design Summary

## Objective

Using https://godoc.org/github.com/shomali11/slacker (recommended by Sasha), I’m writing a Golang process that functions as a slackbot to post new github issues. The slackbot will listen for mentions (or private messages) and take a command and two string parameters in slack. The first string parameter is a repository, the second string parameter is text. This will become a new github issue, and the bot will reply with link or failure. When starting the process, specify a slack key, a github organization or user, and a newline-delimited list of authorized slack users.

### Example in CLI:

```
$: ./issuebot --org=gravitational --auth=./userlist --slack-token=SLACK_TOKEN --github-token=GITHUB_TOKEN
...connecting to slack… success
...pinging github api… success
...looking for userlist… success
...looking for organization… success
(other basic logging)
```

<kbd>^C</kbd>: will give option to reload authed users or quit cleanly (ie stop accepting commands and finish whatever is in queue)  
<kbd>^C ^C</kbd>: get out as fast as possible (kill, not necessarily waiting for network to finish)

### Example in Slack:

**ayjay_t:**  
@issuebot new teleport Won’t compile on atari  
**issuebot:**  
Success: https://github.com/nlopes/slack/issues/460  

Or

**ayjay_t:**  
/msg @issuebot new issuebot Need to authenticate with SSO  
**issuebot:**  
Failure: Network error | No/Bad Repo | etc  

(if I can do custom commands I guess I will? `/issue new` makes sense if it’s easy- will assess)

## Basic Architecture:

The program will be a simple loop (after an init) that listens on slack and then dispatches to linux. The slack uses a callback like structure, so hopefully this is asynchronous.
You can see the *PROJECT_MANAGEMENT.md* file for more information about the approach- a lot of it is specific to the api's being used.

## Misc

[Gravitational Issue Template](https://github.com/gravitational/teleport/blob/master/docs/ISSUE_TEMPLATE.md)

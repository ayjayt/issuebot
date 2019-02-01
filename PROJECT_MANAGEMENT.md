*NB: Since the project relies on previous unused third party libraries, it is a bit "exploratory" and the time estimates and task list are a best-guess. The sheets-app I use tracks more data and lets me track the accuracy of my time-estimates and identify problem areas... but it's not scaled for multiple people. This is the process I've used to bill/estimate, usually multiplying by 3 or requiring an "exploratory" phase to get a more accurate estimate.*


# Task list

| Task | Time Estimate | Status |
| -------------:|:-------------:| ----:|
| Early research | -- | Done @ 2:00 |
| PM timing | 30 min | Done @ 0:35 |
| Look at commenting documentation | 30 min max | Pending |
| Look at slack documentation (the one Sasha recommend shama11?) | 30 min max | Pending |
| Look at github docs (google repo) | 30 min max | Pending |
| Look at x/oauth (apparently both slack and github use oauth) | 45 min budgeted | Pending |
| Set up a development/testing environment | 30 min budgeted | Pending |
| Set up dep manager (vgo? godep? go mod? whatever is easier) | 45 min budgeted | Pending |
| Catch signals to set a healthy exit | 10 min | Pending |
| Set up flag module to take in flags | 15 min | Pending |
| Error checking on flags/sanitizing/sanity check | Basics only, 15 min | Pending |
| Input file to array for user auth | 25 min | Pending |
| Write function to sanity test slack connection | 45 min | Pending |
| Write function to sanity test github conneciton + org validity | 45 min | Pending |
| Set up retry function so we can reconnect after init | 30 min | Pending |
| Set up retry function to retry connection @ decreasing intervals | SKIP for PoC | Pending |
| Write callback to get info from slack (and print to log) | 30 min | Pending |
| Write function to get issue onto github w/ proper returns | 1 hr | Pending |
| Write function to respond to slack |  30 min | Pending |
| Write success unittest for internal functions | 1 hr | Pending |
| Write common error unittests for internal functions | SKIP for PoC | Pending |
| Write fuzzers for internal functions | SKIP for PoC |  Pending |
| Total | 13:05 Estimated | 2:35 Completed |

Comment as you go along in godoc-friendly style

Unit tests as you go along to not double up on debugging/unit testing

# Things we're not doing:


* I like putting APIs/interfaces into their own corners but here it's all in one loop. I like doing that because it ensures all APIs are agnostic to other APIs. The key here would be creating a datastructure that represents a command/info and writing shims that translate that datastructure for w/e API is writing it/reading it.
* The auth is based on a list and that is really inflexible- the slack API would be better served to allow the user to write a custom function as the "validator" function that takes relevant parameters (use info, command requested, w/e) and returns a bool: true (user is authed for this operation) or false.
* There is pretty poor error reporting/observability, not that it's really important for such a small app but it helps to put thought into what you're going to do to make it convenient to fix, so I'd put thought there.

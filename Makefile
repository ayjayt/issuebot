.PHONY: build tidy clean test

build:
	@echo "make build:"
	go build -o build/issuebot

clean:
	@echo "make clean:"
	-rm -r build/*

test:
	@echo "make test (expand this w/ hello):"
	go test

# Add files here that you want to be checked before building w/ tools in "tidy" target below
FILES := main.go main_test.go flags.go flags_test.go slack.go slack_test.go github.go github_test.go

tidy:$(FILES)
	gofmt -w $?
	# TODO: govet?



# TODO: build, install, clean, test
# TODO: 

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
FILES := *.go

tidy:$(FILES)
	gofmt -w $?
	# TODO: govet?



# TODO: build, install, clean, test
# TODO: 

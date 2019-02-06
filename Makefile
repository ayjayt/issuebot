.PHONY: build tidy clean

build:
	@echo "make build:"
	go build -o build/issuebot

clean:
	@echo "make clean:"
	-rm -r build/*

# Add files here that you want to be checked before building w/ tools in "tidy" target below
FILES := main.go 

tidy:$(FILES)
	gofmt -w $?
	# TODO: govet?



# TODO: build, install, clean, test
# TODO: 

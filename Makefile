GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=txl
MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

build:
	$(GOBUILD) -o $(BINARY_NAME)

clean:
	$(GOCLEAN)

test:
	$(GOTEST) -v ./...

example: build
	go get github.com/go-sql-driver/mysql
	./$(BINARY_NAME) $(MAKEFILE_DIR)_example || :
	go mod tidy

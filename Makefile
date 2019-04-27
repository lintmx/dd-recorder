# Basic go commands
GOCMD=go
GOGEN=$(GOCMD) generate
GOINS=$(GOCMD) install
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVER=$(GOCMD) version
GOBIN=./bin

# Project
NAME=dd-recorder
VERSION=$(shell git describe --tags || echo "testing version" )

# Go Build ldflags
LDFLAGS=-ldflags "-w -s -X 'main.Name=$(NAME)' -X 'main.Version=$(VERSION)' -X 'main.Build=`date`' -X 'main.GoVersion=`$(GOVER)`'"

all: mkdir
	$(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)\

mkdir:
	@mkdir -p $(GOBIN)

clean:
	@rm -rf $(GOBIN)

configuration:
	@cp config.yml.example $(GOBIN)/config.yml

build-linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-linux-386 recorder.go

release-linux-386: build-linux-386
	@tar zcf $(GOBIN)/$(NAME)-linux-386.tar.gz -C $(GOBIN) $(NAME)-linux-386 config.yml
	@rm -rf $(GOBIN)/$(NAME)-linux-386

build-linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-linux-amd64 recorder.go

release-linux-amd64: build-linux-amd64
	@tar zcf $(GOBIN)/$(NAME)-linux-amd64.tar.gz -C $(GOBIN) $(NAME)-linux-amd64 config.yml
	@rm -rf $(GOBIN)/$(NAME)-linux-amd64

build-darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-darwin-amd64 recorder.go

release-darwin-amd64: build-darwin-amd64
	@tar zcf $(GOBIN)/$(NAME)-darwin-amd64.tar.gz -C $(GOBIN) $(NAME)-darwin-amd64 config.yml
	@rm -rf $(GOBIN)/$(NAME)-darwin-amd64

build-windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-windows-386.exe recorder.go

release-windows-386: build-windows-386
	@zip -j $(GOBIN)/$(NAME)-windows-386.zip $(GOBIN)/config.yml $(GOBIN)/$(NAME)-windows-386.exe
	@rm -rf $(GOBIN)/$(NAME)-windows-386.exe

build-windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-windows-amd64.exe recorder.go

release-windows-amd64: build-windows-amd64
	@zip -j $(GOBIN)/$(NAME)-windows-amd64.zip $(GOBIN)/config.yml $(GOBIN)/$(NAME)-windows-amd64.exe
	@rm -rf $(GOBIN)/$(NAME)-windows-amd64.exe

release: clean \
	mkdir \
	configuration \
	release-linux-386 \
	release-linux-amd64 \
	release-darwin-amd64 \
	release-windows-386 \
	release-windows-amd64
	@rm -rf $(GOBIN)/config.yml
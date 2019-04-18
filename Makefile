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

all:
	@mkdir -p $(GOBIN)
	$(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)

clean:
	@rm -rf $(GOBIN)

build-linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-linux-386 recorder.go

release-linux-386: build-linux-386
	@tar zcf $(GOBIN)/$(NAME)-linux-386.tar.gz config.yml -C $(GOBIN) $(NAME)-linux-386

build-linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-linux-amd64 recorder.go

release-linux-amd64: build-linux-amd64
	@tar zcf $(GOBIN)/$(NAME)-linux-amd64.tar.gz config.yml -C $(GOBIN) $(NAME)-linux-amd64

build-darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-darwin-amd64 recorder.go

release-darwin-amd64: build-darwin-amd64
	@tar zcf $(GOBIN)/$(NAME)-darwin-amd64.tar.gz config.yml -C $(GOBIN) $(NAME)-darwin-amd64

build-windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-windows-386.exe recorder.go

release-windows-386: build-windows-386
	@zip -j $(GOBIN)/$(NAME)-windows-386.zip config.yml $(GOBIN)/$(NAME)-windows-386.exe

build-windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) ${LDFLAGS} -o $(GOBIN)/$(NAME)-windows-amd64.exe recorder.go

release-windows-amd64: build-windows-amd64
	@zip -j $(GOBIN)/$(NAME)-windows-amd64.zip config.yml $(GOBIN)/$(NAME)-windows-amd64.exe

release: release-linux-386 \
	release-linux-amd64 \
	release-darwin-amd64 \
	release-windows-386 \
	release-windows-amd64
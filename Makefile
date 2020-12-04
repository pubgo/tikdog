Project=github.com/pubgo/tikdog
GoPath=$(shell go env GOPATH)
Version=$(shell git tag --sort=committerdate | tail -n 1)
GoROOT=$(shell go env GOROOT)
BuildTime=$(shell date "+%F %T")
CommitID=$(shell git rev-parse HEAD)
GO := GO111MODULE=on go

LDFLAGS += -X "${Project}/version.GoROOT=${GoROOT}"
LDFLAGS += -X "${Project}/version.BuildTime=${BuildTime}"
LDFLAGS += -X "${Project}/version.GoPath=${GoPath}"
LDFLAGS += -X "${Project}/version.CommitID=${CommitID}"
LDFLAGS += -X "${Project}/version.Project=${Project}"
LDFLAGS += -X "${Project}/version.Version=${Version:-v0.0.1}"

.PHONY: build
build:
	go build -ldflags '${LDFLAGS}' -mod vendor -v -o main main.go

.PHONY: install
install:
	@go install -ldflags '${LDFLAGS}' -mod vendor -v .

.PHONY: test
test:
	@go test -race -v ./... -cover

.PHONY: tag_list
tag_list:
	@git tag -n --sort=committerdate | tee | tail -n 5

.PHONY: release
release:
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS) -s -w' -race -v -o bin/darwin/tikdog
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS) -s -w' -race -v -o bin/linux/tikdog
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS) -s -w' -race -v -o bin/windows/tikdog.exe

# statik -src=assets/build/  -include=*.html,*.js,*.json,*.css,*.png,*.svg,*.ico,*.ttf -f
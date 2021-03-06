GO_REPO_ROOT := /go/src/github.com/sarendsen/dokku-limit
BUILD_IMAGE := golang:1.9.1

.PHONY: build-in-docker build clean

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/set subcommands/defaults subcommands/set-defaults

build-in-docker: clean
	docker run --rm \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make build" || exit $$?

build: commands subcommands hooks

commands: **/**/commands.go
	go build $(GO_ARGS) -o commands src/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: src/subcommands/%.go
	go build $(GO_ARGS) -o $@ $<

hooks:
	go build $(GO_ARGS) -o docker-args-deploy src/hooks/docker-args-deploy.go
	go build $(GO_ARGS) -o pre-deploy src/hooks/pre-deploy.go

clean:
	rm -rf commands subcommands docker-args-deploy pre-deploy

src-clean:
	rm -rf src Makefile
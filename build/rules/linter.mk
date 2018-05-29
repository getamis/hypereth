ifeq ($(HOST_OS), Darwin)
LINTER_IGNORE_DIRS := $(subst ./,,$(shell find . -name '.linterignore' -print0 | xargs -0 -n1 dirname | sort -u))
else
LINTER_IGNORE_DIRS := $(subst ./,,$(shell find . -name '.linterignore' -printf '%h\n' | sort -u))
endif

lint: linter-tools
	@$(GOBIN)/gometalinter.v2 --vendor --skip="$(LINTER_IGNORE_DIRS)" --disable-all --deadline=10m --enable=vet --enable=gofmt --enable=misspell --enable=goconst --enable=unconvert --enable=gosimple --min-occurrences=6 ./...

linter-tools: metalinter
	@$(GOBIN)/gometalinter.v2 --install

metalinter:
	@go install gopkg.in/alecthomas/gometalinter.v2

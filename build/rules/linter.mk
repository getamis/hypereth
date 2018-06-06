ifeq ($(HOST_OS), Darwin)
LINTER_IGNORE_DIRS := $(subst ./,,$(shell find . -name '.linterignore' -print0 | xargs -0 -n1 dirname | sort -u))
else
LINTER_IGNORE_DIRS := $(subst ./,,$(shell find . -name '.linterignore' -print | sort -u))
endif

LINTER_PKG := gopkg.in/alecthomas/gometalinter.v2

lint: linter-tools
	@$(GOBIN)/$(notdir $(LINTER_PKG)) --vendor --skip="$(LINTER_IGNORE_DIRS)" --disable-all --deadline=10m --enable=vet --enable=gofmt --enable=misspell --enable=goconst --enable=unconvert --enable=gosimple --min-occurrences=6 ./...

linter-tools: metalinter
	@$(GOBIN)/$(notdir $(LINTER_PKG)) --install

metalinter:
	@go get -u $(LINTER_PKG)

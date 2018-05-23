PHONY += all docker test clean

CURDIR := $(shell pwd)
GOBIN = $(shell pwd)/build/bin
APP_NAME := $(notdir $(CURDIR))
DOCKER_REPOSITORY := quay.io/amis
DOCKER_IMAGE := $(DOCKER_REPOSITORY)/$(APP_NAME)
ifeq ($(REV),)
REV := $(shell git rev-parse --short HEAD 2> /dev/null)
endif

TARGETS := $(sort $(notdir $(wildcard ./cmd/*)))
PHONY += $(TARGETS) $(APP_NAME)

all: $(TARGETS)

.SECONDEXPANSION:
$(TARGETS): $(addprefix $(GOBIN)/,$$@)

$(GOBIN):
	@mkdir -p $@

$(GOBIN)/%: $(GOBIN) FORCE
	@go build -v -o $@ ./cmd/$(notdir $@)
	@echo "Done building."
	@echo "Run \"$(subst $(CURDIR),.,$@)\" to launch $(notdir $@)."

coverage.txt:
	@touch $@

test: coverage.txt FORCE
	@for d in `go list ./... | grep -v vendor`; do		\
		go test -v -coverprofile=profile.out -covermode=atomic $$d;	\
		if [ $$? -eq 0 ]; then						\
			echo "\033[32mPASS\033[0m:\t$$d";			\
			if [ -f profile.out ]; then				\
				cat profile.out >> coverage.txt;		\
				rm profile.out;					\
			fi							\
		else								\
			echo "\033[31mFAIL\033[0m:\t$$d";			\
			exit -1;						\
		fi								\
	done;

docker:
	@docker build -t $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV) .
	@docker tag $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV) $(DOCKER_REPOSITORY)/$(APP_NAME):latest

docker.push:
	@docker push $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV)
	@docker push $(DOCKER_REPOSITORY)/$(APP_NAME):latest

clean:
	rm -fr $(GOBIN)/*

PHONY: help
help:
	@echo  'Generic targets:'
	@echo  '  all                           - Build all targets marked with [*]'
	@echo  '* hypereth                      - Build hypereth'
	@echo  ''
	@echo  'Docker targets:'
	@echo  '  docker                        - Build hypereth docker image'
	@echo  '  docker.push                   - Push hypereth docker image to quay.io'
	@echo  ''
	@echo  'Test targets:'
	@echo  '  test                          - Run all unit tests'
	@echo  ''
	@echo  'Cleaning targets:'
	@echo  '  clean                         - Remove built executables'
	@echo  ''
	@echo  'Execute "make" or "make all" to build all targets marked with [*] '
	@echo  'For further info see the ./README.md file'

.PHONY: $(PHONY)

.PHONY: FORCE
FORCE:

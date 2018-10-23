PHONY += all docker test clean

CURDIR := $(shell pwd)
BUILD_DIR := $(CURDIR)/build
GOBIN = $(BUILD_DIR)/bin
APP_NAME := $(notdir $(CURDIR))
TARGETS := $(sort $(notdir $(wildcard ./cmd/*)))
PHONY += $(TARGETS) $(APP_NAME)

HOST_OS := $(shell uname -s)

export GOBIN
export PATH := $(GOBIN):$(PATH)

all: $(TARGETS)

.SECONDEXPANSION:
$(TARGETS): $(addprefix $(GOBIN)/,$$@)
$(GOBIN):
	@mkdir -p $@

$(GOBIN)/%: $(GOBIN) FORCE
	@go build -v -o $@ ./cmd/$(notdir $@)
	@echo "Done building."
	@echo "Run \"$(subst $(CURDIR),.,$@)\" to launch $(notdir $@)."

include $(wildcard $(BUILD_DIR)/rules/*.mk)

clean:
	rm -fr $(GOBIN)/*

PHONY: help
help:
	@echo  'Generic targets:'
	@echo  '  all                           - Build all targets marked with [*]'
	@echo  '* metrics-exporter              - Build metrics-exporter'
	@echo  ''
	@echo  'Docker targets:'
	@echo  '  docker                        - Build hypereth docker image'
	@echo  '  docker.push                   - Push hypereth docker image to quay.io'
	@echo  ''
	@echo  'Test targets:'
	@echo  '  lint                          - Run linters'
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

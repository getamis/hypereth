DOCKER_REPOSITORY := quay.io/amis
DOCKER_IMAGE := $(DOCKER_REPOSITORY)/$(APP_NAME)
ifeq ($(REV),)
REV := $(shell git rev-parse --short HEAD 2> /dev/null)
endif

docker:
	@docker build -t $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV) .
	@docker tag $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV) $(DOCKER_REPOSITORY)/$(APP_NAME):latest

docker.push:
	@docker push $(DOCKER_REPOSITORY)/$(APP_NAME):$(REV)
	@docker push $(DOCKER_REPOSITORY)/$(APP_NAME):latest

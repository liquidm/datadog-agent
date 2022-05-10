PROJECT_NAME = datadog-agent

PROJECT_REV = $(shell git rev-parse HEAD)
PROJECT_IMAGE = registry.lqm.io/$(PROJECT_NAME):$(PROJECT_REV)
PROJECT_TEST_IMAGE = registry.lqm.io/$(PROJECT_NAME)-test:$(PROJECT_REV)

TMP_ARTIFACT_DIR := $(shell mktemp -d)
TMP_ARTIFACT_FILE := $(shell mktemp)

TEST ?= false

default: build

all: build

build:
	go build -o $(PROJECT_NAME)

test:
	go test -race ./...

image:
	docker build --build-arg TEST -t $(PROJECT_IMAGE) .

artifact: image
	$(eval CID := $(shell docker create $(PROJECT_IMAGE)))
	docker cp $(CID):/var/app/$(PROJECT_NAME) $(TMP_ARTIFACT_DIR)
	docker rm $(CID)
	tar -zcf $(TMP_ARTIFACT_FILE) -C $(TMP_ARTIFACT_DIR) $(PROJECT_NAME)

publish-image: image
	docker push $(PROJECT_IMAGE)

publish-artifact: artifact
	gsutil cp $(TMP_ARTIFACT_FILE) gs://lqm-artifact-storage/$(PROJECT_NAME)/$(PROJECT_REV)

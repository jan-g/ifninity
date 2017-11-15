GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
# GOPATH can take multiple values - only grab the first as that's where go get puts stuff
GOLINTPATH =$(shell echo $$GOPATH | sed -e 's/:.*//')/bin/golint
# Just builds
all: myday build

dep: glide.yaml
	glide install --strip-vendor

dep-up:
	glide up --strip-vendor

vet: $(GOFILES)
	go vet $(GOPACKAGES)

lint: $(GOFILES)
	OK=0; for pkg in $(GOPACKAGES) ; do   echo Running golint $$pkg ;  $(GOLINTPATH) $$pkg  || OK=1 ;  done ; exit $$OK

test: $(shell find . -name *.go)
	go test -v $(GOPACKAGES)

build: $(GOFILES)
	go build -o fn-mock-vista

fmt: $(GOFILES)
	gofmt -w -s $(GOFILES)

myday: test lint vet

COMPLETER_DIR := $(realpath $(dir $(firstword $(MAKEFILE_LIST))))
CONTAINER_COMPLETER_DIR := /go/src/github.com/jan-g/ifninity

IMAGE_REPO_USER ?= ioctl
IMAGE_NAME ?= fn-mock-vista
IMAGE_VERSION ?= latest
IMAGE_FULL = $(IMAGE_REPO_USER)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_LATEST = $(IMAGE_REPO_USER)/$(IMAGE_NAME):latest

docker-pull-image-funcy-go:
	docker pull funcy/go:dev

docker-test: docker-pull-image-funcy-go
	docker run --rm -it -v $(COMPLETER_DIR):$(CONTAINER_COMPLETER_DIR) -w $(CONTAINER_COMPLETER_DIR) -e CGO_ENABLED=1 funcy/go:dev sh -c 'go test -v $$(go list ./...  | grep -v /vendor/)'

docker-build: docker-test docker-pull-image-funcy-go
	docker run --rm -it -v $(COMPLETER_DIR):$(CONTAINER_COMPLETER_DIR) -w $(CONTAINER_COMPLETER_DIR) -e CGO_ENABLED=1 funcy/go:dev go build -o fn-mock-vista
	docker build -t $(IMAGE_FULL) -f $(COMPLETER_DIR)/Dockerfile $(COMPLETER_DIR)

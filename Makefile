BINARY := dash-backend
VERSION := 0.1.0
COMMIT = $(shell git rev-parse HEAD)
SHA = $(shell git rev-parse --short HEAD)
CURR_DIR = $(shell pwd)
CURR_DIR_WIN = $(shell cd)
BIN_DIR = $(CURR_DIR)/build
BIN_DIR_WIN = $(CURR_DIR_WIN)/build
export GO111MODULE = on

BRANCH := $(shell bash -c 'if [ "$$TRAVIS_PULL_REQUEST" == "false" ]; then echo $$TRAVIS_BRANCH; else echo $$TRAVIS_PULL_REQUEST_BRANCH; fi')

# Set BRANCH when running make manually
ifeq ($(BRANCH),)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif

# Setup the -ldflags option to pass vars defined here to app vars
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.branch=${BRANCH}"

PKGS = $(shell go list ./...)

PLATFORMS := windows linux darwin
os = $(word 1, $@)

ifeq ($(BRANCH),develop)
        DOCKER_IMAGE_REPO := dash-backend
else
        DOCKER_IMAGE_REPO := dash-backend-dev
endif

all: install build
.PHONY: all


install:
ifeq ($(OS),Windows_NT) 
	setup_env.bat
else
	./setup_env.sh
endif
.PHONY: install


genproto:
ifeq ($(OS),Windows_NT) 
	scripts\win\genproto.bat
else
	./scripts/genproto.sh
endif
.PHONY: genproto


build:
ifeq ($(OS),Windows_NT)
	go build ${LDFLAGS} -o $(BIN_DIR_WIN)/$(BINARY).exe
else
	go build ${LDFLAGS} -o $(BIN_DIR)/$(BINARY)
endif
.PHONY: build


tidy:
	go mod tidy
.PHONY: tidy


$(PLATFORMS): genproto
ifeq ($(OS),Windows_NT)
	set GOOS=$(os)&&set GOARCH=amd64&&go build ${LDFLAGS} -o $(CURR_DIR)/$(BINARY)
else
	GOOS=$(os) GOARCH=amd64 go build ${LDFLAGS} -o $(CURR_DIR)/$(BINARY)
endif
.PHONY: $(PLATFORMS)


arm6: genproto
	GOOS=linux GOARCH=arm GOARM=6 go build ${LDFLAGS} -o $(CURR_DIR)/$(BINARY)
.PHONY: pi


test: genproto
	ulimit -n 9999; go test -timeout 0 -p 1 ./...
.PHONY: test


test-no-app-test: genproto
	ulimit -n 9999; go test -timeout 0 -p 1 -tags exclude_app_test ./...
.PHONY: test


test-only-app-test: genproto
	ulimit -n 9999; go test -timeout 0 -p 1 -v -tags !exclude_app_test ./cmd/node
.PHONY: test


test-tidy:
	# Working directory must be clean, or this test would be destructive
	git diff --quiet || (echo "\033[0;31mWorking directory not clean!\033[0m" && exit 1)
	# We expect `go mod tidy` not to change anything, the test should fail otherwise
	make tidy
	git diff --exit-code || (git checkout . && exit 1)
.PHONY: test-tidy


test-fmt:
	git diff --quiet || (echo "\033[0;31mWorking directory not clean!\033[0m" && exit 1)
	# We expect `go fmt` not to change anything, the test should fail otherwise
	go fmt ./...
	git diff --exit-code || (git checkout . && exit 1)
.PHONY: test-fmt

lint:
	golint --set_exit_status ./...
	go vet ./...
.PHONY: lint


cover:
	@echo "mode: count" > cover-all.out
	@$(foreach pkg,$(PKGS),\
		go test -coverprofile=cover.out -covermode=count $(pkg);\
		tail -n +2 cover.out >> cover-all.out;)
	go tool cover -html=cover-all.out
.PHONY: cover


dockerbuild-go:
	docker build -t $(DOCKER_IMAGE_REPO):$(BRANCH) .
.PHONY: dockerbuild-go


dockerpush: dockerbuild-go
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	docker tag $(DOCKER_IMAGE_REPO):$(BRANCH) spacemeshos/$(DOCKER_IMAGE_REPO):$(BRANCH)
	docker push spacemeshos/$(DOCKER_IMAGE_REPO):$(BRANCH)

ifeq ($(BRANCH),develop)
	docker tag $(DOCKER_IMAGE_REPO):$(BRANCH) spacemeshos/$(DOCKER_IMAGE_REPO):$(SHA)
	docker push spacemeshos/$(DOCKER_IMAGE_REPO):$(SHA)
endif
.PHONY: dockerpush

ifdef TEST
DELIM=::
endif

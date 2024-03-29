PROJECT_DIR = $(shell pwd)
PROJECT_BUILD = ${PROJECT_DIR}/build
PROJECT_BIN = ${PROJECT_DIR}/bin
$(shell [ -f bin ] || mkdir -p ${PROJECT_BIN})

BINARY_NAME = chatyx-backend
BRANCH_NAME = $(shell git name-rev --name-only HEAD)
COMMIT_HASH = $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP = $(shell date +"%Y-%m-%d:T%H:%M:%S")
PACKAGE = github.com/Chatyx/backend
LDFLAGS = -X '${PACKAGE}/version.BranchName=${BRANCH_NAME}' \
  -X '${PACKAGE}/version.CommitHash=${COMMIT_HASH}' \
  -X '${PACKAGE}/version.BuildTimestamp=${BUILD_TIMESTAMP}' \

GOLANGCI_LINT = ${PROJECT_BIN}/golangci-lint
SWAG = ${PROJECT_BIN}/swag
MIGRATE = ${PROJECT_BIN}/migrate
MOCKERY = ${PROJECT_BIN}/mockery

export PATH := ${PROJECT_BIN}:${PATH}

all: clean lint test test.integration build

.PHONY: .install-linter
.install-linter:
	[ -f ${GOLANGCI_LINT} ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${PROJECT_BIN} v1.54.2

.PHONY: .install-swagger
.install-swagger:
	[ -f ${SWAG} ] || GOBIN=${PROJECT_BIN} go install github.com/swaggo/swag/cmd/swag@v1.16.2

.PHONY: .install-migrate
.install-migrate:
	[ -f ${MIGRATE} ] || GOBIN=${PROJECT_BIN} go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

.PHONY: .install-mockery
.install-mockery:
	[-f ${MOCKERY} ] || GOBIN=${PROJECT_BIN} go install github.com/vektra/mockery/v2@v2.40.1

.PHONY: lint
lint: .install-linter
	${GOLANGCI_LINT} --version
	${GOLANGCI_LINT} linters
	${GOLANGCI_LINT} run -v

.PHONY: lint.fix
lint.fix:
	${GOLANGCI_LINT} run --fix

.PHONY: swagger
swagger: .install-swagger
	${SWAG} init -g ./internal/transport/http/entry.go -o ./api

.PHONY: build
build:
	go build -ldflags="${LDFLAGS}" -o ${PROJECT_BUILD}/${BINARY_NAME} ./cmd/${BINARY_NAME}

.PHONY: generate
generate: .install-mockery
	go generate ./...

.PHONY: test
test:
	go test -count=1 -v -coverprofile=cover.out ./internal/... ./pkg/...
	go tool cover -func=cover.out

.PHONY: test.integration
test.integration:
	docker-compose -f ./test/docker-compose.yaml up -d
	go test -count=1 -v ./test/... -run TestAppTestSuite || true
	docker-compose -f ./test/docker-compose.yaml down

# usage: make migration NAME="{migration_name}"
.PHONY: migration
migration: .install-migrate
	${MIGRATE} create -ext sql -dir ./db/migrations ${NAME}

.PHONY: migrate.up
migrate.up: .install-migrate
	${MIGRATE} -path=./db/migrations/ \
        -database postgres://chatyx_user:chatyx_password@localhost:5432/chatyx_db?sslmode=disable up

.PHONY: migrate.down
migrate.down: .install-migrate
	${MIGRATE} -path=./db/migrations/ \
        -database postgres://chatyx_user:chatyx_password@localhost:5432/chatyx_db?sslmode=disable down 1

.PHONY: clean
clean:
	rm -rf ${PROJECT_BUILD} || true
	rm -rf ${PROJECT_BIN} || true
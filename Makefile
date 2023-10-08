PROJECT_DIR = $(shell pwd)
PROJECT_BUILD = ${PROJECT_DIR}/build
PROJECT_BIN = ${PROJECT_DIR}/bin
$(shell [ -f bin ] || mkdir -p ${PROJECT_BIN})

BINARY_NAME = chatyx
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

all: clean lint test.unit test.integration build

.PHONY: .install-linter
.install-linter:
	[ -f ${GOLANGCI_LINT} ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${PROJECT_BIN} v1.54.2

.PHONY: .install-swagger
.install-swagger:
	[ -f ${SWAG} ] || GOBIN=${PROJECT_BIN} go install github.com/swaggo/swag/cmd/swag@v1.16.2

.PHONY: .install-migrate
.install-migrate:
	[ -f ${MIGRATE} ] || GOBIN=${PROJECT_BIN} go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2

.PHONY: lint
lint: .install-linter
	${GOLANGCI_LINT} --version
	${GOLANGCI_LINT} linters
	${GOLANGCI_LINT} run -v

.PHONY: swagger
swagger: .install-swagger
	${SWAG} init -g ./internal/app/app.go

.PHONY: build
build:
	go build -ldflags="${LDFLAGS}" -o ${PROJECT_BUILD}/${BINARY_NAME} ./cmd/chatyx-backend

.PHONY: generate
generate:
	go generate ./...

.PHONY: infra
infra:
	docker-compose up --remove-orphan postgres redis

.PHONY: test.unit
test.unit:
	go test -tags=unit -v -coverprofile=cover.out ./...
	go tool cover -func=cover.out

.PHONY: test.integration
test.integration: infra
	go test -tags=integration -v ./test/... || true

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
	rm -rf ${PROJECT_BIN} || true
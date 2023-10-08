BINARY_NAME=chatyx
BUILD_DIR=./build
TEST_DIR=${BUILD_DIR}/tests
BRANCH_NAME="$(shell git name-rev --name-only HEAD)"
COMMIT_HASH="$(shell git rev-parse --short HEAD)"
BUILD_TIMESTAMP=$(shell date +"%Y-%m-%d:T%H:%M:%S")
PACKAGE_SDK=github.com/Chatyx/backend
LDFLAGS= -X '${PACKAGE_SDK}/version.BranchName=${BRANCH_NAME}' \
  -X '${PACKAGE_SDK}/version.CommitHash=${COMMIT_HASH}' \
  -X '${PACKAGE_SDK}/version.BuildTimestamp=${BUILD_TIMESTAMP}' \

all: clean lint test.unit test.integration build

lint:
	golangci-lint --version
	golangci-lint linters
	golangci-lint run -v

build:
	go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/chatyx-backend

swagger:
	swag init -g ./internal/app/app.go

generate:
	go generate ./...

infra:
	docker-compose up --remove-orphan postgres redis

test.unit:
	go test -tags=unit -v -coverprofile=cover.out ./...
	go tool cover -func=cover.out

test.integration: infra
	go test -tags=integration -v ./test/... || true

migrations:
	migrate create -ext sql -dir ./db/migrations $(NAME)

migrate.up:
	migrate -path=./db/migrations/ \
        -database postgres://chatyx_user:chatyx_password@localhost:5432/chatyx_db?sslmode=disable up

migrate.down:
	migrate -path=./db/migrations/ \
        -database postgres://chatyx_user:chatyx_password@localhost:5432/chatyx_db?sslmode=disable down 1

clean:
	rm -rf ${BUILD_DIR} || true
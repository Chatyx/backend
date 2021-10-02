#!make
include .env

lint:
	golangci-lint run

build:
	go build -o ./build/scht-backend ./cmd/app/main.go

generate:
	go generate ./...

test.unit:
	go test -v -coverprofile=cover.out ./... && go tool cover -func=cover.out

postgres:
	docker stop scht-postgres || true
	docker run --rm --detach --name=scht-postgres \
		--env POSTGRES_USER=${SCHT_PG_USERNAME} \
		--env POSTGRES_PASSWORD=${SCHT_PG_PASSWORD} \
		--env POSTGRES_DB=${SCHT_PG_DATABASE} \
		--publish ${SCHT_PG_PORT}:5432 postgres:12.1

redis:
	docker stop scht-redis || true
	docker run --rm --detach --name=scht-redis \
		--publish ${SCHT_REDIS_PORT}:6379 redis:6.2.5

create-migration:
	docker run --rm -v ${PWD}/db/migrations:/migrations \
		migrate/migrate create -ext sql -dir /migrations -seq $(NAME)

migrate:
	docker run --rm -v ${PWD}/db/migrations:/migrations \
		--network host migrate/migrate \
        -path=/migrations/ \
        -database postgres://${SCHT_PG_USERNAME}:${SCHT_PG_PASSWORD}@${SCHT_PG_HOST}:${SCHT_PG_PORT}/${SCHT_PG_DATABASE}?sslmode=disable up

downgrade:
	docker run --rm -v ${PWD}/db/migrations:/migrations \
    	--network host migrate/migrate \
        -path=/migrations/ \
        -database postgres://${SCHT_PG_USERNAME}:${SCHT_PG_PASSWORD}@${SCHT_PG_HOST}:${SCHT_PG_PORT}/${SCHT_PG_DATABASE}?sslmode=disable down 1
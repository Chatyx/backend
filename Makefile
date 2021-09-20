#!make
include .env

lint:
	golangci-lint run

build:
	go build -o ./builds/scht-backend ./cmd/app/main.go

postgres:
	docker stop scht-postgres || true
	docker run --rm --detach --name=scht-postgres \
		--env POSTGRES_USER=scht_user \
		--env POSTGRES_PASSWORD=scht_password \
		--env POSTGRES_DB=scht_db \
		--publish 5432:5432 postgres:12.1

echo:
	echo ${SCHT_PG_DATABASE}
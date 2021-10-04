lint:
	golangci-lint run

build:
	go build -o ./build/scht-backend ./cmd/app/main.go

swagger:
	swag init -g ./internal/app/app.go

generate:
	go generate ./...

test.unit:
	go test -v -coverprofile=cover.out ./... && go tool cover -func=cover.out

infrastructure.dev:
	docker-compose -f docker-compose.dev.yml down
	docker-compose -f docker-compose.dev.yml up -d

migrations:
	docker run --rm -v ${PWD}/internal/db/migrations:/migrations \
		migrate/migrate create -ext sql -dir /migrations -seq $(NAME)

migrate.up:
	docker run --rm -v ${PWD}/internal/db/migrations:/migrations \
		--network host migrate/migrate \
        -path=/migrations/ \
        -database postgres://scht_user:scht_password@localhost:5432/scht_db?sslmode=disable up

migrate.down:
	docker run --rm -v ${PWD}/internal/db/migrations:/migrations \
    	--network host migrate/migrate \
        -path=/migrations/ \
        -database postgres://scht_user:scht_password@localhost:5432/scht_db?sslmode=disable down 1
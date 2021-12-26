BACKEND_BIN = ./build/scht-backend
MIGRATE_BIN = ./build/migrate

lint:
	golangci-lint run

build: clean $(BACKEND_BIN) $(MIGRATE_BIN)

$(BACKEND_BIN):
	go build -o $(BACKEND_BIN) ./cmd/app/main.go

$(MIGRATE_BIN):
	go build -o $(MIGRATE_BIN) ./cmd/migrate/main.go

swagger:
	swag init -g ./internal/app/app.go

generate:
	go generate ./...

infrastructure.dev:
	docker-compose down
	docker-compose up --remove-orphan postgres redis

infrastructure.test:
	docker-compose -f docker-compose.test.yml down
	docker-compose -f docker-compose.test.yml up -d

test.unit:
	go test -tags=unit -v -coverprofile=cover.out ./...
	go tool cover -func=cover.out

test.integration: infrastructure.test
	go test -tags=integration -v ./test/... || true
	docker-compose -f docker-compose.test.yml down

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

clean:
	rm -rf ./build || true
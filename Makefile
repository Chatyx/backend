lint:
	golangci-lint run

build:
	go build -o ./builds/scht-backend ./cmd/app/main.go

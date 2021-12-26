FROM golang:1.17 as builder

# Define build environment variables
ENV GOOS linux
ENV CGO_ENABLED  0

# Add a working directory
WORKDIR /scht-backend

# Copy files with defined dependencies to the working directory
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy application files to the working directory
COPY . ./

# Build application and other cli utilities
RUN go build -o app ./cmd/app/main.go && \
    go build -o migrate ./cmd/migrate/main.go


FROM alpine:3.15

# Add a working directory
WORKDIR /scht-backend

# Copy built binaries, configs and migrations from builder to the /scht-backend directory
COPY --from=builder /scht-backend/internal/db/migrations ./internal/db/migrations
COPY --from=builder /scht-backend/configs ./configs
COPY --from=builder /scht-backend/app ./
COPY --from=builder /scht-backend/migrate ./

# Define volumes
VOLUME ./configs

# Expose ports
EXPOSE 8000 8080

# Execute built binary
CMD ./app

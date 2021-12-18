FROM golang:1.16 as builder

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

# Build application
RUN go build -o app ./cmd/app/main.go


FROM alpine:3.15

# Add a working directory
WORKDIR /scht-backend

# Copy built binary from builder to the /scht-backend directory
COPY --from=builder /scht-backend/app ./
COPY --from=builder /scht-backend/configs ./configs

VOLUME ./configs
VOLUME ./logs

# Expose ports
EXPOSE 8000 8080

# Execute build binary
CMD ./app

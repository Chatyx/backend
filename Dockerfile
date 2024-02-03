FROM golang:1.21 as builder

# Define build environment variables
ENV GOOS linux
ENV CGO_ENABLED 0

# Add a working directory
WORKDIR /chatyx-backend

# Copy files with defined dependencies to the working directory
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy application files to the working directory
COPY . ./

# Build application and tools
RUN make build PROJECT_BUILD=. && make .install-migrate


FROM alpine:3.18

# Add a working directory
WORKDIR /chatyx-backend

# Copy built binaries, configs and migrations from builder to the /chatyx-backend directory
COPY --from=builder /chatyx-backend/db/migrations ./db/migrations
COPY --from=builder /chatyx-backend/configs ./configs
COPY --from=builder /chatyx-backend/chatyx-backend ./
COPY --from=builder /chatyx-backend/bin/migrate ./

# Define volumes
VOLUME ./configs

# Expose ports
EXPOSE 8080 8081

# Execute built binary
CMD ./chatyx-backend

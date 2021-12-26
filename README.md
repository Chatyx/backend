# SCHT Backend

## Description

SCHT Backend is a simple chat implementation in GO.

## Architecture
TODO

## Configuration

Basic configuration defined as a single YAML-file:

```yaml
is_debug: false
domain: localhost
listen:
  api:
    type: port
    bind_ip: 0.0.0.0
    bind_port: 8000
  chat:
    type: port
    bind_ip: 0.0.0.0
    bind_port: 8080
auth:
  sign_key: xNN3f9M9vZLtqHJwX2wtTbCBMpR # env: SCHT_AUTH_SIGN_KEY
  access_token_ttl: 15                  # 15 minutes
  refresh_token_ttl: 43200              # 30 days
postgres:
  host: 127.0.0.1                       # env: SCHT_PG_HOST
  port: 5432                            # env: SCHT_PG_PORT
  database: scht_db                     # env: SCHT_PG_DATABASE
  username: scht_user                   # env: SCHT_PG_USERNAME
  password: scht_password               # env: SCHT_PG_PASSWORD
  max_conn_attempts: 3
  failed_conn_delay: 5                  # 5 seconds
redis:
  host: 127.0.0.1                       # env: SCHT_REDIS_HOST
  port: 6379                            # env: SCHT_REDIS_PORT
  username: default                     # env: SCHT_REDIS_USERNAME
  password: ""                          # env: SCHT_REDIS_PASSWORD
  max_conn_attempts: 3
  failed_conn_delay: 5                  # 5 seconds
logging:
  level: debug
  filepath: ./logs/all.log
  rotate: true
  max_size: 100                         # 100 MB
  max_backups: 5                        # 5 max files
cors:
  allowed_origins:
    - 127.0.0.1:3000
    - localhost:3000
  max_age: 600                          # 10 minutes
```

You can configure part of parameters with environment variables 
(like these: `SCHT_AUTH_SIGN_KEY`, `SCHT_PG_PASSWORD`, etc).

## Installation

### Using single docker container

In the beginning, you need to create a working directory and go to it.

```bash
$ mkdir ~/scht-backend && cd ~/scht-backend
```

Then, you should create configuration file (you can copy default config from `Configuration` section above).

```bash
$ mkdir configs && vi configs/main.yml
```

and change hostnames for Postgres and Redis like this:

```yaml
postgres:
  host: docker
  port: 5432
  ...

redis:
  host: docker
  port: 6379
  ...
```

At least you need to run single docker container and apply migrations:

```bash
$ docker run --rm --add-host=docker:<YOUR_DOCKER_HOST_IP> \
--volume=$(PWD)/configs:/scht-backend/configs \
--publish=8000:8000 --detach \
--name=scht-backend mortalis/scht-backend:latest

$ docker exec scht-backend ./migrate
```

For getting the IP address of docker host you can use the following shell command
(working on Debian):

```bash
$ ip route show 0.0.0.0/0 | grep -Eo 'via \S+' | awk '{ print $2 }'
```

### Using docker-compose

In the beginning, you need to create a working directory and go to it.

```bash
$ mkdir ~/scht-backend && cd ~/scht-backend
```

Since the application uses Postgres and Redis as external infrastructure services, it was a good solution to create
`docker-compose.yml` file in the working directory and define all required services. Create `docker-compose.yml` and
paste this content.

```yaml
version: "3.9"

services:
  backend:
    image: mortalis/scht-backend:latest
    container_name: scht-backend
    ports:
      - "8000:8000"
      - "8080:8080"
    volumes:
      - ./configs:/scht-backend/configs
    depends_on:
      - postgres
      - redis
    restart: always
    networks:
      - scht-backend-network

  postgres:
    image: postgres:12.1
    container_name: scht-postgres
    environment:
      - POSTGRES_DB=scht_db
      - POSTGRES_USER=scht_user
      - POSTGRES_PASSWORD=scht_password
    volumes:
      - ./.volumes/postgres/data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: always
    networks:
      - scht-backend-network

  redis:
    image: redis:6.2.5
    container_name: scht-redis
    volumes:
      - ./.volumes/redis/data:/data
    ports:
      - "6379:6379"
    restart: always
    networks:
      - scht-backend-network

networks:
  scht-backend-network:
```

After this you should create configuration file (you can copy default config from `Configuration` section above).

```bash
$ mkdir configs && vi configs/main.yml
```

and change hostnames for Postgres and Redis like this:

```yaml
postgres:
  host: postgres
  port: 5432
  ...

redis:
  host: redis
  port: 6379
  ...
```

At least you need to run containers and apply migrations:

```bash
$ docker-compose up --detach
$ docker exec scht-backend ./migrate
```

### Manually building from source code

In the beginning, you need to clone this repository:

```bash
$ git clone git@github.com:Mort4lis/scht-backend.git
# or
$ git clone https://github.com/Mort4lis/scht-backend.git
```

Next go to the cloned directory and execute:

```bash
$ cd scht-backend

$ make build
# or
$ go build -o ./build/scht-backend ./cmd/app/main.go && \
  go build -o ./build/migrate ./cmd/migrate/main.go
```

Finally, run the application:

```bash
$ ./build/scht-backend
# or
$ ./build/scht-backend --config=<PATH_TO_THE_CONFIG>
```

Don't forget to apply migrations:

```bash
$ ./build/migrate
```

## How to use

If you need to substitute config, you can do it very simply:

```bash
$ ./scht-backend --config=<PATH_TO_THE_CONFIG>
```

After running the application you can use REST API for sign up users, authenticate them, creating chats, adding members
to the created chat and so on. See swagger documentation `http://localhost:8000/docs`
for more details.



Also, available to you WebSocket API for sending and receiving messages in the real time.
(by default at `ws://localhost:8080`). For getting that you should generate code for your 
language from [proto file](./internal/encoding/proto/message.proto) and use `CreateMessageDTO` to
send messages and `Message` to handle received message.
See [documentation](https://developers.google.com/protocol-buffers) for more details.
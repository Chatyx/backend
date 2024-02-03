# Chatyx Backend

![ci](https://github.com/Mort4lis/scht-backend/actions/workflows/main.yml/badge.svg)
![license](https://img.shields.io/github/license/Chatyx/backend)
![go-version](https://img.shields.io/github/go-mod/go-version/Chatyx/backend)
![docker-pulls](https://img.shields.io/docker/pulls/mortalis/chatyx-backend)
![code-size](https://img.shields.io/github/languages/code-size/Chatyx/backend)
![total-lines](https://img.shields.io/tokei/lines/github/Chatyx/backend)

## 📖 Description

Chatyx backend is an MVP monolith message service implemented in Go. The project will evolve 
towards a microservice architecture. The target design is described in [this page](./docs/system_design/README.md).

## 🚀 Features

Already done:
* ✅ Support groups and dialogs
* ✅ Support for sending text messages both via REST and Websocket
* ✅ Add and remove participants for group chats
* ✅ Participants can leave from group chats
* ✅ Block partners in dialogs 

Not done yet:
* ❌ View unread messages
* ❌ Show online/offline statuses of users, as well as when the user was last online
* ❌ Notifications if user isn't online
* ❌ Support cross-device synchronization

## 🔧 Installation

### Using single docker container

```bash
docker run --rm --volume=$(PWD)/configs:/chatyx-backend/configs \
  --publish=8080:8080 --publish=8081:8081 --detach \
  --name=chatyx-backend mortalis/chatyx-backend:latest

# Apply migrations
docker exec chatyx-backend ./migrate -path=./db/migrations/ -database 'postgres://<POSTGRES_USER>:<POSTGRES_PASSWORD>@<POSTGRES_HOST>:<POSTGRES_PORT>/<POSTGRES_DB>?sslmode=disable' up
```

### Manually building from source code

```bash
git clone git@github.com:Chatyx/backend.git chatyx-backend && cd chatyx-backend
make build

# Apply migrations
./bin/migrate -path=./db/migrations/ -database 'postgres://<POSTGRES_USER>:<POSTGRES_PASSWORD>@<POSTGRES_HOST>:<POSTGRES_PORT>/<POSTGRES_DB>?sslmode=disable' up

# Run the application
./build/chatyx-backend --config=<PATH_TO_THE_CONFIG>
```

## ⚙️ Configuration

Basic configuration defined as a single [YAML file](./configs/config.yaml):

You can configure part of parameters with environment variables like these: 
`POSTGRES_USER`, `POSTGRES_PASSWORD`, etc. The full list of supported environment variables
are described in a config file after comment prefix `# env: `.

To run the application with substituted config you should perform:

```bash
$ ./chatyx-backend --config=<PATH_TO_THE_CONFIG>
```

## 📈 How to use

After running the application you can use REST API for creating groups and dialogs, adding participants
sending messages and so on. See swagger documentation `http://localhost:8080/swagger` for more details.

Also, available to you WebSocket API for sending and receiving messages in the real time.
(by default at `ws://localhost:8081`). For getting that you should generate code for your 
language from [proto file](./internal/transport/websocket/model/message.proto) and use `MessageCreate` to
send messages and `Message` to receive message.

See [documentation](https://developers.google.com/protocol-buffers) for more details.
# Chatyx Backend

![ci](https://github.com/Mort4lis/scht-backend/actions/workflows/main.yml/badge.svg)
![license](https://img.shields.io/github/license/Chatyx/backend)
![go-version](https://img.shields.io/github/go-mod/go-version/Chatyx/backend)
![docker-pulls](https://img.shields.io/docker/pulls/mortalis/scht-backend)
![code-size](https://img.shields.io/github/languages/code-size/Chatyx/backend)
![total-lines](https://img.shields.io/tokei/lines/github/Chatyx/backend)

## Description

Chatyx backend is a simple MVP chat application implemented in GO.

## Design application

TBD

## Configuration

Basic configuration defined as a single [YAML-file](./configs/config.yaml):

You can configure part of parameters with environment variables 
(like these: `CHATYX_POSTGRES_USER`, `CHATYX_POSTGRES_PASSWORD`, etc).

## Installation

TBD

## How to use

If you need to substitute config, you can do it very simply:

```bash
$ ./chatyx --config=<PATH_TO_THE_CONFIG>
```

After running the application you can use REST API for creating groups and dialogs, adding participants
sending messages and so on. See swagger documentation `http://localhost:8080/docs` for more details.

Also, available to you WebSocket API for sending and receiving messages in the real time.
(by default at `ws://localhost:8081`). For getting that you should generate code for your 
language from [proto file](./internal/encoding/proto/message.proto) and use `CreateMessageDTO` to
send messages and handle received message.

See [documentation](https://developers.google.com/protocol-buffers) for more details.
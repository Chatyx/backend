BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username   VARCHAR(50) UNIQUE  NOT NULL,
    password   VARCHAR(255)        NOT NULL,
    first_name VARCHAR(50)         NULL,
    last_name  VARCHAR(50)         NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    birth_date DATE                NULL,
    department VARCHAR(255)        NULL,
    is_deleted BOOLEAN          DEFAULT false,
    created_at TIMESTAMPTZ      DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS chats
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT         NULL,
    creator_id  UUID         NOT NULL,
    created_at  TIMESTAMPTZ      DEFAULT current_timestamp,
    updated_at  TIMESTAMPTZ,

    CONSTRAINT creator_fk FOREIGN KEY (creator_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS users_chats
(
    user_id UUID NOT NULL,
    chat_id UUID NOT NULL,

    PRIMARY KEY (user_id, chat_id),
    CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT chat_fk FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text       TEXT NOT NULL,
    author_id  UUID NOT NULL,
    chat_id    UUID NOT NULL,
    created_at TIMESTAMPTZ      DEFAULT current_timestamp,

    CONSTRAINT author_fk FOREIGN KEY (author_id) REFERENCES users (id),
    CONSTRAINT chat_fk FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

COMMIT;
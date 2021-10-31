BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS chat_member_status_list
(
    id   SMALLINT PRIMARY KEY NOT NULL,
    name VARCHAR(50)          NOT NULL
);

CREATE TABLE IF NOT EXISTS message_action_list
(
    id   SMALLINT PRIMARY KEY NOT NULL,
    name VARCHAR(50)          NOT NULL
);

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

CREATE TABLE IF NOT EXISTS chat_members
(
    user_id   UUID     NOT NULL,
    chat_id   UUID     NOT NULL,
    status_id SMALLINT NOT NULL DEFAULT 1,

    PRIMARY KEY (user_id, chat_id),
    CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT chat_fk FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE,
    CONSTRAINT status_fk FOREIGN KEY (status_id) REFERENCES chat_member_status_list (id)
);

CREATE TABLE IF NOT EXISTS messages
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id  SMALLINT NOT NULL,
    text       TEXT     NOT NULL,
    sender_id  UUID     NOT NULL,
    chat_id    UUID     NOT NULL,
    created_at TIMESTAMPTZ      DEFAULT current_timestamp,

    CONSTRAINT author_fk FOREIGN KEY (sender_id) REFERENCES users (id),
    CONSTRAINT chat_fk FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE,
    CONSTRAINT action_fk FOREIGN KEY (action_id) REFERENCES message_action_list (id)
);

INSERT INTO chat_member_status_list (id, name)
VALUES (1, 'In chat'),
       (2, 'Left'),
       (3, 'Kicked');

INSERT INTO message_action_list (id, name)
VALUES (1, 'Send message to chat'),
       (2, 'User joined to chat'),
       (3, 'User left from chat'),
       (4, 'Admin kicked this user');

COMMIT;
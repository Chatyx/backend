BEGIN;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'chat_type') THEN
            CREATE TYPE chat_type as ENUM (
                'dialog',
                'group');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'group_participant_status') THEN
            CREATE TYPE group_participant_status as ENUM (
                'joined',
                'left',
                'kicked');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'content_type') THEN
            CREATE TYPE content_type as ENUM (
                'text',
                'image');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS users
(
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(50) UNIQUE       NOT NULL,
    pwd_hash   VARCHAR(255)             NOT NULL,
    email      VARCHAR(255) UNIQUE      NOT NULL,
    first_name VARCHAR(50)              NULL,
    last_name  VARCHAR(50)              NULL,
    birth_date DATE                     NULL,
    bio        VARCHAR(10000)           NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NULL,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS chats
(
    id          BIGSERIAL PRIMARY KEY,
    type        chat_type                NOT NULL,
    uname       VARCHAR(255) UNIQUE      NULL,
    name        VARCHAR(255)             NULL,
    description VARCHAR(10000)           NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS group_participants
(
    chat_id  BIGINT                   NOT NULL
        REFERENCES chats (id) ON DELETE CASCADE,
    user_id  BIGINT                   NOT NULL
        REFERENCES users (id),
    status   group_participant_status NOT NULL DEFAULT 'joined',
    is_admin BOOLEAN                           DEFAULT FALSE,

    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE IF NOT EXISTS dialog_participants
(
    chat_id    BIGINT NOT NULL
        REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL
        REFERENCES users (id),
    is_blocked BOOLEAN DEFAULT FALSE,

    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages
(
    id           BIGSERIAL PRIMARY KEY,
    sender_id    BIGINT                   NOT NULL
        REFERENCES users (id),
    chat_id      BIGINT                   NOT NULL
        REFERENCES chats (id) ON DELETE CASCADE,
    content      VARCHAR(2000)            NOT NULL,
    content_type content_type             NOT NULL,
    is_service   BOOLEAN DEFAULT FALSE,
    sent_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    delivered_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS read_messages
(
    chat_id    BIGINT NOT NULL
        REFERENCES chats (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL
        REFERENCES users (id),
    message_id BIGINT NOT NULL
        REFERENCES messages (id),

    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE IF NOT EXISTS last_messages
(
    chat_id    BIGINT PRIMARY KEY
        REFERENCES chats (id) ON DELETE CASCADE,
    message_id BIGINT NOT NULL
        REFERENCES messages (id)
);

COMMIT;
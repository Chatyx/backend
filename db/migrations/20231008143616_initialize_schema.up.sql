BEGIN;

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
                'file',
                'service');
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
    is_active  BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NULL,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS groups
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255)             NOT NULL,
    description VARCHAR(10000)           NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS group_participants
(
    id       BIGSERIAL PRIMARY KEY,
    user_id  BIGINT                   NOT NULL
        REFERENCES users (id),
    group_id BIGINT                   NOT NULL
        REFERENCES groups (id) ON DELETE CASCADE,
    status   group_participant_status NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,

    UNIQUE (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS conversations
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS conversation_participants
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL
        REFERENCES users (id),
    conv_id    BIGINT NOT NULL
        REFERENCES conversations (id) ON DELETE CASCADE,
    is_blocked BOOLEAN DEFAULT FALSE,

    UNIQUE (user_id, conv_id)
);

CREATE TABLE IF NOT EXISTS messages
(
    id           BIGSERIAL PRIMARY KEY,
    sender_id    BIGINT                   NOT NULL
        REFERENCES users (id),
    group_id     BIGINT                   NULL
        REFERENCES groups (id),
    conv_id      BIGINT                   NULL
        REFERENCES conversations (id),
    content      VARCHAR(100000)          NOT NULL,
    content_type content_type             NOT NULL,
    sent_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    delivered_at TIMESTAMP WITH TIME ZONE,
    seen_at      TIMESTAMP WITH TIME ZONE
);

COMMIT;
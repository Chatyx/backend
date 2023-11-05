BEGIN;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS conversation_participants;

DROP TABLE IF EXISTS conversations;

DROP TABLE IF EXISTS group_participants;

DROP TABLE IF EXISTS groups;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS content_type;

DROP TYPE IF EXISTS group_participant_status;

COMMIT;
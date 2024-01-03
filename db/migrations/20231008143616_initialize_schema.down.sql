BEGIN;

DROP TABLE IF EXISTS last_messages;

DROP TABLE IF EXISTS read_messages;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS dialog_participants;

DROP TABLE IF EXISTS group_participants;

DROP TABLE IF EXISTS chats;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS content_type;

DROP TYPE IF EXISTS group_participant_status;

DROP TYPE IF EXISTS chat_type;

COMMIT;
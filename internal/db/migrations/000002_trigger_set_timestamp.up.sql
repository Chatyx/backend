CREATE OR REPLACE FUNCTION trigger_set_timestamp() RETURNS TRIGGER AS
$trigger_set_timestamp$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$trigger_set_timestamp$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp
    BEFORE UPDATE
    ON chats
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE INDEX IF NOT EXISTS dialog_participants__user_id__idx
    ON dialog_participants (user_id);

CREATE INDEX IF NOT EXISTS group_participants__user_id__idx
    ON group_participants (user_id);

CREATE INDEX IF NOT EXISTS messages__chat_id_chat_type__idx
    ON messages (chat_id, chat_type);

CREATE INDEX IF NOT EXISTS messages__sent_at__idx
    ON public.messages (sent_at);
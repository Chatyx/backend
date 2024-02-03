// Replication:
// - master-slave (async)
// - replication factor 2

Enum chat_type {
  "dialog"
  "group"
}

Enum group_participant_status {
  "joined"
  "left"
  "kicked"
}

// Sharding:
// - key based by id

Table chats {
  id bigint [pk, increment]
  type chat_type [not null]
  name varchar(255)
  description varchar(10000)
  created_at timestamp [not null]
  updated_at timestamp
}

// Sharding:
// - key based by chat_id

Table group_participants {
  chat_id bigint [not null]
  user_id bigint [not null]
  status group_participant_status [not null]
  is_admin boolean [default: false]

  Indexes {
    (chat_id, user_id) [pk]
  }
}

Ref: group_participants.chat_id > chats.id [delete: cascade]

// Sharding:
// - key based by chat_id

Table dialog_participants {
  chat_id bigint [not null]
  user_id bigint [not null]
  is_blocked boolean [default: false]

  Indexes {
    (chat_id, user_id) [pk]
  }
}

Ref: dialog_participants.chat_id > chats.id [delete: cascade]
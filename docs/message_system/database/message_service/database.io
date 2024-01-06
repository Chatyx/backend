// Replication:
// - master-less (async)
// - replication factor 2

Enum content_type {
  "text"
  "image"
}

// Sharding:
// - key based by chat_id
//
// Partitioning within the shard:
// - range based by sent_at (per 1 month)

Table messages {
  id bigint [pk, increment]
  sender_id bigint [not null] // user_id
  chat_id bigint [not null]
  content varchar(2000) [not null]
  content_type content_type [not null]
  is_service boolean [default: false]
  sent_at timestamp [not null]
  delivered_at timestamp
}

// Sharding:
// - key based by chat_id

Table read_messages {
  chat_id bigint [not null]
  user_id bigint [not null]
  message_id bigint [not null]

  Indexes {
    (chat_id, user_id) [pk]
  }
}

Ref: read_messages.message_id > messages.id [delete: cascade]

// Sharding:
// - key based by chat_id

Table last_messages {
  chat_id bigint [pk]
  message_id bigint [not null]
}

Ref: last_messages.message_id > messages.id [delete: cascade]
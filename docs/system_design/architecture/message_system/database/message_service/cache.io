// Replication:
// - master-slave (async)
// - replication factor 2
//
// Sharding:
// - key based by chat_id

Table messages {
  chat_id bigint [pk]
  messages list
}
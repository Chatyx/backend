// Replication:
// - master-slave (async)
// - replication factor 2
//
// Sharding:
// - key based by user_id

Table presences {
  user_id bigint [pk]
  last_seen timestamp
}
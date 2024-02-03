package redis

type Session struct {
	UserID      string `redis:"user_id"`
	Fingerprint string `redis:"fingerprint"`
	IP          string `redis:"ip"`
	ExpiresAt   int64  `redis:"expires_at"`
	CreatedAt   int64  `redis:"created_at"`
}

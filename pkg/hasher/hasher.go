package hasher

//go:generate mockgen -source=hasher.go -destination=mocks/mock.go

type PasswordHasher interface {
	Hash(password string) (string, error)
	CompareHashAndPassword(hash, password string) bool
}

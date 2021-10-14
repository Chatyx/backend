package server

type Server interface {
	Run() error
	Stop() error
}

type Config struct {
	ServerName string
	ListenType string
	BindIP     string
	BindPort   int
}

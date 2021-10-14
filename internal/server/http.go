package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type HttpServerWrapper struct {
	*http.Server
	cfg    Config
	logger logging.Logger
}

func NewHttpServerWrapper(cfg Config, base *http.Server) Server {
	return &HttpServerWrapper{
		Server: base,
		cfg:    cfg,
		logger: logging.GetLogger(),
	}
}

func (s *HttpServerWrapper) Run() error {
	var (
		lis    net.Listener
		lisErr error
	)

	switch s.cfg.ListenType {
	case "sock":
		baseDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			s.logger.WithError(err).Errorf("[%s]: Failed to get the base directory", s.cfg.ServerName)
			return errors.New("failed to get the base directory")
		}

		sockPath := path.Join(baseDir, s.getSockName()+".sock")

		s.logger.Infof("[%s]: Bind server to unix socket %s", s.cfg.ServerName, sockPath)
		lis, lisErr = net.Listen("unix", sockPath)
	case "port":
		addr := s.getAddr()

		s.logger.Infof("[%s]: Bind server to tcp %s", s.cfg.ServerName, addr)
		lis, lisErr = net.Listen("tcp", addr)
	}

	if lisErr != nil {
		s.logger.WithError(lisErr).Errorf("[%s]: Failed to listen %s", s.cfg.ServerName, s.cfg.ListenType)
		return fmt.Errorf("failed to listen %s", s.cfg.ListenType)
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				s.logger.Infof("[%s]: Server shutdown", s.cfg.ServerName)
			default:
				s.logger.Fatalf("[%s]: %v", s.cfg.ServerName, err)
			}
		}
	}()

	return nil
}

func (s *HttpServerWrapper) getAddr() string {
	return fmt.Sprintf("%s:%d", s.cfg.BindIP, s.cfg.BindPort)
}

func (s *HttpServerWrapper) getSockName() string {
	sockName := strings.SplitN(s.cfg.ServerName, " ", 2)[0]
	return strings.ToLower(sockName)
}

func (s *HttpServerWrapper) Stop() error {
	return s.Server.Close()
}

package http

import (
	"context"
	"fmt"
	"github.com/teneta-io/dcp/internal/config"
	"github.com/teneta-io/dcp/internal/service"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	ctx         context.Context
	logger      *zap.Logger
	server      *http.Server
	taskService *service.TaskService
}

func New(ctx context.Context, cfg *config.ServerConfig, logger *zap.Logger, taskService *service.TaskService) *Server {
	return &Server{
		ctx:    ctx,
		logger: logger.Named("HTTP"),
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			ReadTimeout:       0,
			ReadHeaderTimeout: 0,
			WriteTimeout:      0,
			IdleTimeout:       0,
			MaxHeaderBytes:    0,
		},
		taskService: taskService,
	}
}

func (s *Server) Run() {
	s.logger.Info("Starting http server...", zap.Any("address", s.server.Addr))

	http.HandleFunc("/healthcheck", s.healthCheckHandler)

	if err := http.ListenAndServe(s.server.Addr, nil); err != nil {
		s.logger.Fatal("server error %v", zap.Error(err))
	}
}

func (s *Server) Shutdown() error {
	s.logger.Info("Shutdown HTTP server...")

	if err := s.server.Shutdown(s.ctx); err != nil {
		s.logger.Fatal("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server stopped.")
	return nil
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

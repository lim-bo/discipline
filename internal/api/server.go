package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/cleanup"
)

type Server struct {
	mx          *chi.Mux
	server      *http.Server
	userService service.UserServiceI
}

type ServicesList struct {
	UserService service.UserServiceI
}

func New(servicesOptions *ServicesList) *Server {
	mx := chi.NewMux()
	return &Server{
		mx: mx,
		server: &http.Server{
			Handler: mx,
		},
		userService: servicesOptions.UserService,
	}
}

func (s *Server) mountEndpoint() {
	s.mx.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.Register)
			r.Post("/login", s.Login)
		})
	})
}

func (s *Server) Run(address string) error {
	s.mountEndpoint()
	s.server.Addr = address
	go func() {
		log.Printf("Server starting on %s", address)
		if err := s.server.ListenAndServe(); err != nil {
			log.Fatal("Server failed: " + err.Error())
		}
	}()
	return s.waitForShutdown()
}

func (s *Server) waitForShutdown() error {
	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-closeCh
	log.Println("Shutting down server...")
	cleanup.CleanUp()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server failed to shutdown: %v", err)
		return err
	}
	log.Println("Server stopped")
	return nil
}

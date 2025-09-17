package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/limbo/discipline/internal/service"
)

type Server struct {
	mx          *chi.Mux
	userService service.UserServiceI
}

type ServicesList struct {
	UserService service.UserServiceI
}

func New(servicesOptions *ServicesList) *Server {
	return &Server{
		mx:          chi.NewMux(),
		userService: servicesOptions.UserService,
	}
}

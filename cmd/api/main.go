// @title Habit-tracker API
// @description API for habit-tracker app "Discipline"
// @BasePath /api/v1
// @schemes http
package main

import (
	"log"

	"github.com/limbo/discipline/internal/api"
	"github.com/limbo/discipline/internal/repository"
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/config"
	jwtservice "github.com/limbo/discipline/pkg/jwt_service"
)

func init() {
	service.InitValidator()
}

func main() {
	cfg := config.New()
	dbCfg := repository.PGCfg{
		Address:  cfg.GetString("POSTGRES_DB_ADDRESS"),
		Username: cfg.GetString("POSTGRES_USER"),
		Password: cfg.GetString("POSTGRES_PASSWORD"),
		DB:       cfg.GetString("POSTGRES_DB"),
	}
	userService := service.NewUserService(repository.NewUsersRepo(&dbCfg))
	habitService := service.NewHabitsService(repository.NewHabitsRepo(&dbCfg))
	serv := api.New(&api.ServicesList{
		UserService:   userService,
		HabitsService: habitService,
		JwtService:    jwtservice.New(cfg.GetString("JWT_SECRET")),
	})
	err := serv.Run(cfg.GetString("API_ADDRESS"))
	if err != nil {
		log.Println("Server error: " + err.Error())
	}
}

// @title Habit-tracker API
// @description API for habit-tracker app "Discipline"
// @BasePath /api/v1
// @schemes http
package main

import (
	"github.com/limbo/discipline/internal/service"
	"github.com/limbo/discipline/pkg/config"
)

func init() {
	service.InitValidator()
}

func main() {
	cfg := config.New()
}

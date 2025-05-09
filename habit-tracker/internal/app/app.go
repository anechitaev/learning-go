package app

import (
	"habit-tracker/internal/config"
	"habit-tracker/internal/handler"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Application struct {
	router *gin.Engine
	cfg    *config.Config
}

func NewApplication() *Application {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := NewDB(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	router := gin.Default()

	registerRoutes(router)

	return &Application{
		router: router,
		cfg:    cfg,
	}
}

func (a *Application) Run() {
	if err := a.router.Run(":" + a.cfg.ServerPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func registerRoutes(router *gin.Engine) {
	router.GET("/ping", handler.PingHandler)
}

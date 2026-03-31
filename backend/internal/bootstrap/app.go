package bootstrap

import (
	"irms/backend/internal/config"
	"irms/backend/internal/middleware"
	"irms/backend/internal/store"

	"github.com/gin-gonic/gin"
)

type App struct {
	Cfg   config.Config
	Store *store.Store
	Engine *gin.Engine
}

func NewApp(cfg config.Config, st *store.Store) *App {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger())

	RegisterRoutes(engine, cfg, st)

	return &App{
		Cfg:   cfg,
		Store: st,
		Engine: engine,
	}
}

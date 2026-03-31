package bootstrap

import (
	swaggerDocs "irms/backend/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func registerSwagger(engine *gin.Engine) {
	swaggerDocs.SwaggerInfo.BasePath = "/api"
	engine.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

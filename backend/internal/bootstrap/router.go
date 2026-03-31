package bootstrap

import (
	"irms/backend/internal/config"
	"irms/backend/internal/controller"
	"irms/backend/internal/middleware"
	"irms/backend/internal/service"
	"irms/backend/internal/store"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine, cfg config.Config, st *store.Store) {
	engine.GET("/api/health", controller.NewHealthController().Health)

	registerSwagger(engine)

	if st != nil {
		controller.NewAuthController(service.NewAuthService(cfg, st.Gorm)).RegisterPublic(engine.Group("/api"))

		authed := engine.Group("/api")
		authed.Use(middleware.AuthRequired(cfg, st.Query))

		controller.NewAuthController(service.NewAuthService(cfg, st.Gorm)).RegisterAuthed(authed)
		controller.NewPermissionController(service.NewPermissionService(st)).Register(authed)
		controller.NewCredentialController(service.NewCredentialService(cfg, st)).Register(authed)

		admin := authed.Group("")
		admin.Use(middleware.SuperAdminOnly())

		controller.NewUserController(service.NewUserService(st)).Register(admin)
		controller.NewUserGroupController(service.NewUserGroupService(st)).Register(admin)
		controller.NewPageController(service.NewPageService(st)).Register(admin)
		controller.NewHostController(service.NewHostService(st)).Register(admin)
		controller.NewServiceController(service.NewServiceService(st)).Register(admin)
		controller.NewEnvironmentController(service.NewEnvironmentService(st)).Register(admin)
		controller.NewLocationController(service.NewLocationService(st)).Register(admin)
		controller.NewHostEnvironmentController(service.NewHostEnvironmentService(st)).Register(admin)
		controller.NewServiceEnvironmentController(service.NewServiceEnvironmentService(st)).Register(admin)
		controller.NewResourceController(service.NewResourceService(st)).Register(admin)
		controller.NewResourceGroupController(service.NewResourceGroupService(st)).Register(admin)
		controller.NewGrantController(service.NewGrantServiceWithStore(st)).Register(admin)
		controller.NewAuditLogController(service.NewAuditService(st.Gorm)).Register(admin)
	}
}

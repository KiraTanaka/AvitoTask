package http

import (
	_ "database/sql"
	"net/http"

	db "avitoTask/internal/db"

	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	TenderHandler TenderHandler
	Routes        *gin.Engine
}

func (route *RouteHandler) InitRoutes(dbModels db.DbModels) {
	route.Routes = gin.Default()
	route.TenderHandler = TenderHandler{tender: dbModels.TenderModel, user: dbModels.UserModel, organization: dbModels.OrganizationModel}
	//route.Routes = gin.Default()

	route.Routes.GET("/", hello)
	routeGroup := route.Routes.Group("/api")
	routeGroup.GET("/ping", ping)

	InitTenderRoutes(routeGroup, &route.TenderHandler)
	//InitBidRoutes(routeGroup)
}
func hello(c *gin.Context) {
	c.JSON(http.StatusOK, "hello")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (route *RouteHandler) Run(serverAddress string) {
	route.Routes.Run(serverAddress)
}

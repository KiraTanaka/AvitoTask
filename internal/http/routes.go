package http

import (
	_ "database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	routes := gin.Default()

	routes.GET("/", hello)
	routeGroup := routes.Group("/api")
	routeGroup.GET("/ping", ping)

	InitTenderRoutes(routeGroup)
	InitBidRoutes(routeGroup)

	return routes

}
func hello(c *gin.Context) {
	c.JSON(http.StatusOK, "hello")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

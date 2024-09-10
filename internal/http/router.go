package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func InitRoutes(conn *sql.DB) *gin.Engine {
	routes := gin.Default()
	routeGroup := routes.Group("/api")

	InitTenderRoutes(routeGroup, conn)
	//InitBidRouters(apiRoutes)

	return routes

}

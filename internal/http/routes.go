package http

import (
	_ "database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func InitRoutes(conn *sqlx.DB) *gin.Engine {
	routes := gin.Default()
	routeGroup := routes.Group("/api")

	InitTenderRoutes(routeGroup, conn)
	//InitBidRouters(apiRoutes)

	return routes

}

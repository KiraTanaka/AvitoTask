package http

import (
	_ "database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func InitRoutes(conn *sqlx.DB) *gin.Engine {
	db = conn
	routes := gin.Default()
	routeGroup := routes.Group("/api")

	InitTenderRoutes(routeGroup)
	//InitBidRouters(apiRoutes)

	return routes

}

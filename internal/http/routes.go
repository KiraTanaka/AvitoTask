package http

import (
	_ "database/sql"
	"net/http"

	db "avitoTask/internal/db"

	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	TenderHandler *TenderHandler
	BidHandler    *BidHandler
	Routes        *gin.Engine
}

func InitRoutes(dbModels *db.DbModels) *RouteHandler {
	route := RouteHandler{
		Routes:        gin.Default(),
		TenderHandler: &TenderHandler{tender: dbModels.TenderModel, user: dbModels.UserModel, organization: dbModels.OrganizationModel},
		BidHandler:    &BidHandler{bid: dbModels.BidModel, tender: dbModels.TenderModel, user: dbModels.UserModel, organization: dbModels.OrganizationModel},
	}

	route.Routes.GET("/", hello)
	routeGroup := route.Routes.Group("/api")
	routeGroup.GET("/ping", ping)

	InitTenderRoutes(routeGroup, route.TenderHandler)
	InitBidRoutes(routeGroup, route.BidHandler)
	return &route
}

func SetDefaultPaginationParamIfEmpty(limit, offset string) (string, string) {
	if limit == "" {
		limit = "5"
	}
	if offset == "" {
		offset = "0"
	}
	return limit, offset
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

package core

import (
	"avitoTask/config"
	"avitoTask/internal/auth"
	"avitoTask/internal/db"
	"avitoTask/internal/http"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	Config *config.Configuration
	Routes http.RouteHandler
	Auth   auth.AuthHandler
}

func NewServer() (*Server, error) {
	server := &Server{}

	var err error
	server.Config, err = config.GetConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	db, err := db.NewDbConnect(server.Config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	server.Routes.InitRoutes(db)
	server.Auth.DbModels = db
	//server.Validator.DbModels = db
	return server, nil
}

func (server *Server) Run() {
	server.Routes.Run(server.Config.ServerAddress)
}

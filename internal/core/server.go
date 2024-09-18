package core

import (
	"avitoTask/config"
	"avitoTask/internal/db"
	"avitoTask/internal/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Config *config.Configuration
	Db     *sqlx.DB
	Routes *gin.Engine
}

func NewServer() (*Server, error) {
	server := &Server{}

	var err error
	server.Config, err = config.GetConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	server.Db, err = db.NewDbConnect(server.Config)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	/*auth.InitAuth(server.Db)
	validator.InitValidator(server.Db)*/
	server.Routes = http.InitRoutes()
	return server, nil
}

func (server *Server)Run()
{
	server.Routes.Run(server.Config.ServerAddress)
}

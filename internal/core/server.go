package core

import (
	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	"avitoTask/internal/db"
	"avitoTask/internal/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" env-required:"true"`
}

type Server struct {
	Config *ServerConfig
	Db     *sqlx.DB
	Routes *gin.Engine
}

func NewServer() (*Server, error) {
	server := &Server{}

	var err error
	server.Config, err = ReadServerConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	server.Db, err = db.NewDbConnect()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	auth.InitAuth(server.Db)
	validator.InitValidator(server.Db)
	server.Routes = http.InitRoutes(server.Db)
	return server, nil
}

func ReadServerConfig() (*ServerConfig, error) {
	var serverConfig ServerConfig
	err := cleanenv.ReadConfig(".env", &serverConfig)
	//err := cleanenv.ReadEnv(&dbconfig)
	if err != nil {
		return nil, fmt.Errorf("Server config error: %w", err)
	}

	log.Info("Reading of the server configuration parameters is completed.")
	return &serverConfig, nil
}

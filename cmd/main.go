package main

import (
	_ "database/sql"

	"avitoTask/internal/core"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func main() {

	var err error
	server, err := core.NewServer()
	if err != nil {
		log.Error(err)
		return
	}

	server.Run()
}

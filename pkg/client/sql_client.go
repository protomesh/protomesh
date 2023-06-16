package client

import (
	"database/sql"

	"github.com/protomesh/go-app"
)

type SqlClient[D any] struct {
	*app.Injector[D]

	DB *sql.DB

	DriverName       app.Config `config:"driver.name,str" default:"postgres" usage:"Driver name to use in the SQL client"`
	ConnectionString app.Config `config:"connection.string,str" usage:"Connection string to connect to SQL database"`
}

func (s *SqlClient[D]) Start() {

	log := s.Log()

	driverName := s.DriverName.StringVal()
	connectionString := s.ConnectionString.StringVal()

	if len(driverName) == 0 || len(connectionString) == 0 {
		log.Panic("Driver name and connection string must be provided for the SQL client")
	}

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		log.Panic("Failed to open connection to database", "error", err, "driverName", driverName)
	}

	s.DB = db

}

func (s *SqlClient[D]) Stop() {

	log := s.Log()

	if err := s.DB.Close(); err != nil {
		log.Error("Error closing database connection", "error", err)
	}

}

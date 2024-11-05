// Service site keeps track of which sites to monitor.
package db

import (
	"encore.dev/storage/sqldb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

// Define a database named 'site', using the database migrations
// in the "./migrations" folder.
var db = sqldb.NewDatabase("site", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// automatically called on startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db.Stdlib(),
	}))
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

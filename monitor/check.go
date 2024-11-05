package monitor

import (
	"context"
	"time"

	"encore.app/db"
	"encore.dev/storage/sqldb"
)

// Check checks a single site.
//
//encore:api public method=POST path=/check/:siteID
func Check(ctx context.Context, siteID int) error {
	site, err := db.Get(ctx, siteID)
	if err != nil {
		return err
	}
	result, err := Ping(ctx, site.URL)
	if err != nil {
		return err
	}

	_, err = connection.Exec(ctx,
		`INSERT INTO checks (site_id, up, checked_at)
		VALUES ($1, $2, $3)`,
		siteID, result.Up, time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}

// Define a database named 'checks', using the database migrations
// in the "./migrations" folder.
var connection = sqldb.NewDatabase("monitor", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

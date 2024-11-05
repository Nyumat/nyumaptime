package monitor

import (
	"context"
    "golang.org/x/sync/errgroup"
	"encore.app/db"
	"encore.dev/storage/sqldb"
	"encore.dev/cron"
)

// Check checks a single site.
//
//encore:api public method=POST path=/check/:siteID
func Check(ctx context.Context, siteID int) error {
	site, err := db.Get(ctx, siteID)
	if err != nil {
		return err
	}

	return check(ctx, site)
}

func check(ctx context.Context, site *db.Site) error {
	result, err := Ping(ctx, site.URL)
	if err != nil {
		return err
	}

	// Publish a Pub/Sub message if the site transitions
	// from up->down or from down->up.
	if err := publishOnTransition(ctx, site, result.Up); err != nil {
		return err
	}

	_, err = connection.Exec(ctx,
		`INSERT INTO checks (site_id, up, checked_at)
		VALUES ($1, $2, NOW())`,
		site.ID, result.Up,
	)

	return err
}


// CheckAll checks all sites.
//
//encore:api public method=POST path=/checkall
func CheckAll(ctx context.Context) error {
	// Get all the tracked sites.
	resp, err := db.List(ctx)
	if err != nil {
		return err
	}

	// Check up to 8 sites concurrently.
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for _, site := range resp.Sites {
		site := site // capture for closure
		g.Go(func() error {
			return check(ctx, site)
		})
	}
	return g.Wait()
}

// Define a database named 'checks', using the database migrations
// in the "./migrations" folder.
var connection = sqldb.NewDatabase("monitor", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var _ = cron.NewJob("check-all", cron.JobConfig{
	Title:    "Check all sites",
	Endpoint: CheckAll,
	Every:    1 * cron.Hour,
})
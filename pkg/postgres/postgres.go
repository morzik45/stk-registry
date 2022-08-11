package postgres

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/morzik45/stk-registry/pkg/config"
	"go.uber.org/zap"
	"strconv"
)

func InitDBx(ctx context.Context, cfg *config.Config, logger *zap.Logger) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("postgres",
		"host="+cfg.Postgres.Host+" port="+strconv.Itoa(cfg.Postgres.Port)+" user="+cfg.Postgres.Username+
			" password="+cfg.Postgres.Password+" dbname="+cfg.Postgres.DBName+" sslmode="+cfg.Postgres.SSLMode)
	if err != nil {
		return nil, errors.New("Postgresql not found!: " + err.Error())
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, errors.New("Postgresql not reply!: " + err.Error())
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return nil, errors.New("Migration error!: " + err.Error())
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	return db, nil
}

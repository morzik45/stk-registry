package postgres

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/morzik45/stk-registry/pkg/config"
	"go.uber.org/zap"
)

type Closable interface {
	Close() error
}

type DB struct {
	logger    *zap.Logger
	DB        *sqlx.DB
	needClose []Closable

	Emails             *Emails
	ErcUpdates         *ErcUpdates
	PersonsFromErc     *PersonsFromERC
	RstkUpdates        *RstkUpdates
	PersonsFromRSTK    *PersonsFromRSTK
	CorrectPersonsData *CorrectPersonsData
	Breakers           *Breakers
}

func NewDB(ctx context.Context, cfg *config.Config, logger *zap.Logger) (db *DB, err error) {
	db = &DB{
		logger: logger,
	}

	db.DB, err = InitDBx(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	db.Emails, err = NewEmails(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.Emails)

	db.ErcUpdates, err = NewErcUpdates(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.ErcUpdates)

	db.PersonsFromErc, err = NewPersonsFromERC(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.PersonsFromErc)

	db.RstkUpdates, err = NewRstkUpdates(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.RstkUpdates)

	db.PersonsFromRSTK, err = NewPersonsFromRSTK(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.PersonsFromRSTK)

	db.CorrectPersonsData, err = NewCorrectPersonsData(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.CorrectPersonsData)

	db.Breakers, err = NewBreakers(ctx, db.DB, logger)
	if err != nil {
		return
	}
	db.needClose = append(db.needClose, db.Breakers)

	return
}

func (db *DB) Close() error {
	for _, c := range db.needClose {
		err := c.Close()
		if err != nil {
			db.logger.Error("Error while closing", zap.Error(err))
		}
	}
	return db.DB.Close()
}

func (db *DB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return db.DB.BeginTxx(ctx, nil)
}

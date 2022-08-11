package postgres

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"time"
)

type PersonFromRSTK struct {
	ID           int       `db:"id"`
	RstkUpdateID int       `db:"rstk_update_id"`
	Snils        string    `db:"snils"`
	Family       string    `db:"family"`
	Name         string    `db:"name"`
	Patronymic   string    `db:"patronymic"`
	Date         time.Time `db:"date"`
	Number       string    `db:"number"`

	Errors pq.StringArray `db:"errors"`
}

type PersonsFromRSTK struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	createMany func(ctx context.Context, persons []PersonFromRSTK, tx *sqlx.Tx) error
}

func NewPersonsFromRSTK(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*PersonsFromRSTK, error) {
	pfr := PersonsFromRSTK{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := pfr.initPersonsFromRSTK(ctxShort)
	if err != nil {
		logger.Error("failed to init personsFromRSTK", zap.Error(err))
		return nil, err
	}
	return &pfr, nil
}

func (pfr *PersonsFromRSTK) Close() error {
	for _, stmt := range pfr.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pfr *PersonsFromRSTK) initPersonsFromRSTK(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	pfr.createMany, stmt, err = pfr.initCreateMany(ctx)
	if err != nil {
		return
	}
	pfr.stmts = append(pfr.stmts, stmt)

	return
}

func (pfr *PersonsFromRSTK) CreateMany(ctx context.Context, persons []PersonFromRSTK, tx *sqlx.Tx) error {
	if pfr.createMany == nil {
		return errors.New("createMany func is not defined")
	}
	return pfr.createMany(ctx, persons, tx)
}

func (pfr *PersonsFromRSTK) initCreateMany(ctx context.Context) (func(ctx context.Context, persons []PersonFromRSTK, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	query := `
		INSERT INTO persons_from_rstk ("rstk_update_id", "snils", "family", "name", "patronymic", "date", "number",
		                               "errors")
		VALUES (:rstk_update_id, :snils, :family, :name, :patronymic, :date, :number, :errors)
	`

	stmt, err := pfr.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, persons []PersonFromRSTK, tx *sqlx.Tx) error {
		// FIXME: С первой попытки не получилось используя NamedStmt вставлять слайс персон в один запрос. Доработать.
		//currentStmt := stmt
		//if tx != nil {
		//	currentStmt = tx.NamedStmtContext(ctx, stmt)
		//}
		_, err = tx.NamedExecContext(ctx, query, persons)
		//_, err = currentStmt.ExecContext(ctx, persons)
		return err
	}, stmt, nil
}

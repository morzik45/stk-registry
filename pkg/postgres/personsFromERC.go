package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"time"
)

type PersonFromERC struct {
	ID          int64     `db:"id"`
	ErcUpdateID int       `db:"erc_update_id"`
	Snils       string    `db:"snils"`
	Birthdate   time.Time `db:"birthdate"`
	Family      string    `db:"family"`
	Name        string    `db:"name"`
	Patronymic  string    `db:"patronymic"`
	Year        int       `db:"year"`
	Semester    int       `db:"semester"`
	Color       string    `db:"color"`
	Count       int       `db:"count"`
	Spent       int       `db:"spent"`
	Date        time.Time `db:"date"`
	CashierID   int       `db:"cashier_id"`
	CashierName string    `db:"cashier_name"`

	Errors pq.StringArray `db:"errors"`
}

type PersonsFromErcForWeb struct {
	Snils       string          `db:"snils" json:"snils"`
	Birthdate   time.Time       `db:"birthdate" json:"birthdate"`
	FullName    string          `db:"full_name" json:"full_name"`
	SaleCoupons json.RawMessage `db:"sale_coupons" json:"sale_coupons"`
}

type PersonsFromERC struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	createMany func(ctx context.Context, persons []PersonFromERC, tx *sqlx.Tx) error
	get        func(ctx context.Context, search string, limit, offset int64) ([]PersonsFromErcForWeb, error)
}

func NewPersonsFromERC(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*PersonsFromERC, error) {
	pfp := PersonsFromERC{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := pfp.initPersonsFromErc(ctxShort)
	if err != nil {
		logger.Error("failed to init personsFromERC", zap.Error(err))
		return nil, err
	}
	return &pfp, nil
}

func (pfp *PersonsFromERC) Close() error {
	for _, stmt := range pfp.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (pfp *PersonsFromERC) initPersonsFromErc(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	pfp.createMany, stmt, err = pfp.initCreateMany(ctx)
	if err != nil {
		return
	}
	pfp.stmts = append(pfp.stmts, stmt)

	pfp.get, stmt, err = pfp.initGet(ctx)
	if err != nil {
		return
	}
	pfp.stmts = append(pfp.stmts, stmt)

	return
}

func (pfp *PersonsFromERC) CreateMany(ctx context.Context, persons []PersonFromERC, tx *sqlx.Tx) error {
	if pfp.createMany == nil {
		return errors.New("createMany func is not defined")
	}
	return pfp.createMany(ctx, persons, tx)
}

func (pfp *PersonsFromERC) initCreateMany(ctx context.Context) (func(ctx context.Context, persons []PersonFromERC, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	query := `
		INSERT INTO persons_from_erc ("erc_update_id", "snils", "birthdate", "family", "name", "patronymic", "year",
		                                   "semester", "color", "count", "spent", "date", "cashier_id", "cashier_name",
		                                   "errors")
		VALUES (:erc_update_id, :snils, :birthdate, :family, :name, :patronymic, :year, :semester, :color, :count,
		        :spent, :date, :cashier_id, :cashier_name, :errors)`

	stmt, err := pfp.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, persons []PersonFromERC, tx *sqlx.Tx) error {
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

func (pfp *PersonsFromERC) Get(ctx context.Context, search string, limit, offset int64) ([]PersonsFromErcForWeb, error) {
	if pfp.get == nil {
		return nil, errors.New("get func is not defined")
	}
	return pfp.get(ctx, search, limit, offset)
}

func (pfp *PersonsFromERC) initGet(ctx context.Context) (func(ctx context.Context, search string, limit, offset int64) ([]PersonsFromErcForWeb, error), *sqlx.NamedStmt, error) {
	query := `
		WITH a AS (SELECT DISTINCT "snils", "birthdate", "family", "name", "patronymic"
				   FROM persons_from_erc
				   WHERE LOWER("family") LIKE '%' || LOWER(:search) || '%'
					  OR LOWER("name") LIKE '%' || LOWER(:search) || '%'
					  OR LOWER("patronymic") LIKE '%' || LOWER(:search) || '%'
					  OR "snils" LIKE '%' || :search || '%'
				   ORDER BY "family"
				   LIMIT :limit OFFSET :offset)
		SELECT a."snils"                                              as "snils",
			   a."birthdate"                                           AS "birthdate",
			   a."family" || ' ' || a."name" || ' ' || a."patronymic" AS "full_name",
			   (SELECT to_json(array_agg(row_to_json(d)))
				FROM (SELECT "id", "count", "date", "color", '(' || "cashier_id" || ') ' || "cashier_name" AS "cashier"
					  FROM persons_from_erc
					  WHERE "snils" = a."snils") d)                   as "sale_coupons"
		FROM a;`

	stmt, err := pfp.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, search string, limit, offset int64) (persons []PersonsFromErcForWeb, err error) {
		if limit == 0 {
			limit = 100
		}
		err = stmt.SelectContext(ctx, &persons, map[string]interface{}{
			"search": search,
			"limit":  limit,
			"offset": offset,
		})
		return
	}, stmt, nil
}

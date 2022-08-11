package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"time"
)

type ArrayError []string

func (eue *ArrayError) Value() (driver.Value, error) {
	return eue.MarshalJSON()
}

func (eue *ArrayError) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return eue.UnmarshalJSON(b)
}

func (eue *ArrayError) MarshalJSON() ([]byte, error) {
	return json.Marshal(eue)
}

func (eue *ArrayError) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, eue); err != nil {
		return err
	}
	return nil
}

type ErcUpdate struct {
	ID      int    `db:"id"`
	EmailID int    `db:"email_id"`
	Name    string `db:"name"`
}

type ErcUpdateInfo struct {
	ID               int64           `db:"id" json:"id"`
	DatetimeReceived time.Time       `db:"datetime_received" json:"datetime_received"`
	DatetimeParsed   time.Time       `db:"datetime_parsed" json:"datetime_parsed"`
	Lines            int             `db:"lines" json:"lines"`
	Incorrect        json.RawMessage `db:"incorrect" json:"incorrect"`
}

type ErcUpdateStats struct {
	Total       int `db:"total" json:"total"`
	Sales       int `db:"sales" json:"sales"`
	Quantity    int `db:"quantity" json:"quantity"`
	Amount      int `db:"amount" json:"amount"`
	Retirees    int `db:"retirees" json:"retirees"`
	UpdatesRSTK int `db:"updates_rstk" json:"updates_rstk"`
	Cards       int `db:"cards" json:"cards"`
}

type ErcUpdateError struct {
	ID        int64          `db:"id" json:"id"`
	Snils     string         `db:"snils" json:"snils"`
	Birthdate string         `db:"birthdate" json:"birthdate"`
	FullName  string         `db:"full_name" json:"full_name"`
	Errors    pq.StringArray `db:"errors" json:"errors"`
}

type ErcUpdates struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	create    func(ctx context.Context, update *ErcUpdate, tx *sqlx.Tx) error
	getInfo   func(ctx context.Context) ([]ErcUpdateInfo, error)
	getStats  func(ctx context.Context) (ErcUpdateStats, error)
	getErrors func(ctx context.Context) ([]ErcUpdateError, error)
}

func NewErcUpdates(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*ErcUpdates, error) {
	eus := ErcUpdates{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := eus.initErcUpdates(ctxShort)
	if err != nil {
		logger.Error("failed to init ercUpdates", zap.Error(err))
		return nil, err
	}
	return &eus, nil
}

func (eus *ErcUpdates) Close() error {
	for _, stmt := range eus.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (eus *ErcUpdates) initErcUpdates(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	eus.create, stmt, err = eus.initCreate(ctx)
	if err != nil {
		return
	}
	eus.stmts = append(eus.stmts, stmt)

	eus.getInfo, stmt, err = eus.initGetInfo(ctx)
	if err != nil {
		return
	}
	eus.stmts = append(eus.stmts, stmt)

	eus.getStats, stmt, err = eus.initGetStats(ctx)
	if err != nil {
		return
	}
	eus.stmts = append(eus.stmts, stmt)

	eus.getErrors, stmt, err = eus.initGetErrors(ctx)
	if err != nil {
		return
	}
	eus.stmts = append(eus.stmts, stmt)

	return
}

func (eus *ErcUpdates) Create(ctx context.Context, update *ErcUpdate, tx *sqlx.Tx) error {
	if eus.create == nil {
		return errors.New("create func is not defined")
	}
	return eus.create(ctx, update, tx)
}

func (eus *ErcUpdates) initCreate(ctx context.Context) (func(ctx context.Context, update *ErcUpdate, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	stmt, err := eus.db.PrepareNamedContext(ctx, `
		INSERT INTO erc_updates ("email_id", "name")
		VALUES (:email_id, :name)
		RETURNING id`)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, update *ErcUpdate, tx *sqlx.Tx) error {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		return currentStmt.GetContext(ctx, &update.ID, *update)
	}, stmt, nil
}

func (eus *ErcUpdates) GetInfo(ctx context.Context) ([]ErcUpdateInfo, error) {
	if eus.getInfo == nil {
		return []ErcUpdateInfo{}, errors.New("getInfo func is not defined")
	}
	return eus.getInfo(ctx)
}

func (eus *ErcUpdates) initGetInfo(ctx context.Context) (func(ctx context.Context) ([]ErcUpdateInfo, error), *sqlx.NamedStmt, error) {
	stmt, err := eus.db.PrepareNamedContext(ctx, `
		SELECT eu.id,
			   e.datetime_received,
			   e.datetime_parsed,
			   COALESCE((SELECT count(*) FROM persons_from_erc AS pfe WHERE pfe.erc_update_id = eu.id), 0) AS "lines",
			   COALESCE((SELECT to_json(array_agg(row_to_json(d)))
				FROM (SELECT "id",
							 pfe."snils"                                                  as "snils",
							 pfe."birthdate"                                               AS "birthdate",
							 pfe."family" || ' ' || pfe."name" || ' ' || pfe."patronymic" AS "full_name",
							 pfe.errors
					  FROM persons_from_erc pfe
					  WHERE pfe."erc_update_id" = eu.id AND pfe.errors IS NOT NULL) d), '[]')    AS "incorrect"
		FROM erc_updates AS eu
				 LEFT JOIN emails e on e.id = eu.email_id;`,
	)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context) (info []ErcUpdateInfo, err error) {
		err = stmt.SelectContext(ctx, &info, map[string]interface{}{})
		return info, err
	}, stmt, nil
}

func (eus *ErcUpdates) GetStats(ctx context.Context) (stats ErcUpdateStats, err error) {
	if eus.getStats == nil {
		return ErcUpdateStats{}, errors.New("getStats func is not defined")
	}
	return eus.getStats(ctx)
}

func (eus *ErcUpdates) initGetStats(ctx context.Context) (func(ctx context.Context) (ErcUpdateStats, error), *sqlx.NamedStmt, error) {
	stmt, err := eus.db.PrepareNamedContext(ctx, `
		SELECT COALESCE((SELECT count(*) FROM "erc_updates"), 0) AS "total",
			   COALESCE((SELECT count(*) FROM "persons_from_erc"), 0) AS "sales",
			   COALESCE((SELECT sum("count") FROM "persons_from_erc"), 0) AS "quantity",
			   COALESCE((SELECT sum("spent") FROM "persons_from_erc"), 0) AS "amount",
			   COALESCE((SELECT count(DISTINCT snils) FROM "persons_from_erc" WHERE snils != ''), 0) AS "retirees",
			   COALESCE((SELECT count(*) FROM "rstk_updates"), 0) AS "updates_rstk",
			   COALESCE((SELECT count(*) FROM "persons_from_rstk"), 0) AS cards;`,
	)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context) (stats ErcUpdateStats, err error) {
		err = stmt.GetContext(ctx, &stats, map[string]interface{}{})
		return stats, err
	}, stmt, nil
}

func (eus *ErcUpdates) GetErrors(ctx context.Context) ([]ErcUpdateError, error) {
	if eus.getErrors == nil {
		return []ErcUpdateError{}, errors.New("getErrors func is not defined")
	}
	return eus.getErrors(ctx)
}

func (eus *ErcUpdates) initGetErrors(ctx context.Context) (func(ctx context.Context) ([]ErcUpdateError, error), *sqlx.NamedStmt, error) {
	stmt, err := eus.db.PrepareNamedContext(ctx, `
		SELECT "id",
			   "snils",
			   "birthdate",
			   "family" || ' ' || "name" || ' ' || "patronymic" AS "full_name",
			   "errors"
		FROM "persons_from_erc"
		WHERE "errors" IS NOT NULL;`,
	)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context) (errors []ErcUpdateError, err error) {
		err = stmt.SelectContext(ctx, &errors, map[string]interface{}{})
		return errors, err
	}, stmt, nil
}

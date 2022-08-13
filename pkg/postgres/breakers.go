package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type Breaker struct {
	ID       int64  `json:"id" db:"id"`
	Snils    string `json:"snils" db:"snils"`
	Checked  bool   `json:"checked" db:"checked"`
	Datetime string `json:"datetime" db:"datetime"`
}

type BreakerView struct {
	Date     string          `json:"date" db:"date"`
	Snils    string          `json:"snils" db:"snils"`
	Name     string          `json:"name" db:"name"`
	Pan      string          `json:"pan" db:"pan"`
	Checked  bool            `json:"checked" db:"checked"`
	Timeline json.RawMessage `json:"timeline" db:"timeline"`
}

type Breakers struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	create  func(ctx context.Context, breaker *Breaker, tx *sqlx.Tx) error
	getView func(ctx context.Context, tx *sqlx.Tx) ([]BreakerView, error)
}

func NewBreakers(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*Breakers, error) {
	br := Breakers{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := br.initBreakers(ctxShort)
	if err != nil {
		logger.Error("failed to init Breakers", zap.Error(err))
		return nil, err
	}
	return &br, nil
}

func (br *Breakers) Close() error {
	for _, stmt := range br.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (br *Breakers) initBreakers(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	br.create, stmt, err = br.initCreate(ctx)
	if err != nil {
		return
	}
	br.stmts = append(br.stmts, stmt)

	br.getView, stmt, err = br.initGetView(ctx)
	if err != nil {
		return
	}
	br.stmts = append(br.stmts, stmt)

	return
}

func (br *Breakers) Create(ctx context.Context, breaker *Breaker, tx *sqlx.Tx) error {
	if br.create == nil {
		return errors.New("create func is not initialized")
	}
	return br.create(ctx, breaker, tx)
}

func (br *Breakers) initCreate(ctx context.Context) (func(ctx context.Context, breaker *Breaker, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	query := `
		INSERT INTO public.breakers (snils, checked)
		VALUES (:snils, :checked)
		RETURNING id, snils, checked, datetime;
	`
	stmt, err := br.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, breaker *Breaker, tx *sqlx.Tx) error {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		return currentStmt.GetContext(ctx, breaker, breaker)
	}, stmt, nil
}

func (br *Breakers) GetView(ctx context.Context, tx *sqlx.Tx) ([]BreakerView, error) {
	if br.getView == nil {
		return nil, errors.New("create func is not initialized")
	}
	return br.getView(ctx, tx)
}

func (br *Breakers) initGetView(ctx context.Context) (func(ctx context.Context, tx *sqlx.Tx) ([]BreakerView, error), *sqlx.NamedStmt, error) {
	query := `
		WITH A AS (SELECT DISTINCT ON (r.number) r.number                                         AS pan,
												 r.date                                           AS date,
												 r.snils                                          AS snils,
												 r.family || ' ' || r.name || ' ' || r.patronymic AS name,
												 COALESCE((SELECT b.checked
														   FROM breakers b
														   WHERE b.snils = r.snils
														   ORDER BY datetime DESC
														   LIMIT 1), FALSE)                       AS checked,
		
												 (SELECT to_json(array_agg(row_to_json(d)))
												  FROM (SELECT r1.date                               AS timestamp,
															   'Активирована карта с №' || r1.number AS content
														FROM persons_from_rstk r1
														WHERE r1.snils = r.snils
														UNION ALL
														SELECT e1.date                             AS timestamp,
															   'Куплено ' || e1.count || 'талонов' AS content
														FROM persons_from_erc e1
														WHERE e1.snils = r.snils
														ORDER BY timestamp) AS d)
																								  AS timeline
				   FROM persons_from_rstk r
							INNER JOIN persons_from_erc e ON r.snils = e.snils
				   ORDER BY r.number DESC)
		SELECT *
		FROM a
		ORDER BY a.date DESC;
	`
	stmt, err := br.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, tx *sqlx.Tx) (views []BreakerView, err error) {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		err = currentStmt.SelectContext(ctx, &views, map[string]interface{}{})
		return views, err
	}, stmt, nil
}

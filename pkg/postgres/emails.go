package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type Email struct {
	ID               int       `db:"id"`
	TypeID           int       `db:"type_id"`
	MessageID        string    `db:"message_id"`
	FromAddress      string    `db:"from_address"`
	DatetimeReceived time.Time `db:"datetime_received"`
	DatetimeParsed   time.Time `db:"datetime_parsed"`
	File             []byte    `db:"file"`
}

type Emails struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	create              func(ctx context.Context, email *Email, tx *sqlx.Tx) error
	getLastReceivedTime func(ctx context.Context, tx *sqlx.Tx) (time.Time, error)
}

func NewEmails(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*Emails, error) {
	es := Emails{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := es.initEmails(ctxShort)
	if err != nil {
		logger.Error("failed to init emails", zap.Error(err))
		return nil, err
	}
	return &es, nil
}

func (es *Emails) Close() error {
	for _, stmt := range es.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (es *Emails) initEmails(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	es.create, stmt, err = es.initCreate(ctx)
	if err != nil {
		return
	}
	es.stmts = append(es.stmts, stmt)

	es.getLastReceivedTime, stmt, err = es.initGetLastReceivedTime(ctx)
	if err != nil {
		return
	}
	es.stmts = append(es.stmts, stmt)

	return
}

func (es *Emails) Create(ctx context.Context, email *Email, tx *sqlx.Tx) error {
	if es.create == nil {
		return errors.New("create func is not defined")
	}
	return es.create(ctx, email, tx)
}

func (es *Emails) initCreate(ctx context.Context) (func(ctx context.Context, email *Email, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	query := `
		INSERT INTO emails (type_id, message_id, from_address, datetime_received, datetime_parsed, file)
		VALUES (:type_id, :message_id, :from_address, :datetime_received, :datetime_parsed, :file)
		RETURNING id;
	`
	stmt, err := es.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare statement %s: %s", query, err.Error())
	}

	return func(ctx context.Context, email *Email, tx *sqlx.Tx) (err error) {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		err = currentStmt.GetContext(ctx, &email.ID, *email)
		return
	}, stmt, nil
}

func (es *Emails) GetLastReceivedTime(ctx context.Context, tx *sqlx.Tx) (time.Time, error) {
	if es.getLastReceivedTime == nil {
		return time.Time{}, errors.New("getLastReceivedTime func is not defined")
	}
	return es.getLastReceivedTime(ctx, tx)
}

func (es *Emails) initGetLastReceivedTime(ctx context.Context) (func(ctx context.Context, tx *sqlx.Tx) (time.Time, error), *sqlx.NamedStmt, error) {
	query := `
		SELECT datetime_received
		FROM emails
		ORDER BY datetime_received DESC
		LIMIT 1;
	`
	stmt, err := es.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare statement %s: %s", query, err.Error())
	}

	return func(ctx context.Context, tx *sqlx.Tx) (dt time.Time, err error) {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		ddt := struct {
			DatetimeReceived time.Time `db:"datetime_received"`
		}{}
		err = currentStmt.GetContext(ctx, &ddt, struct{}{})
		return ddt.DatetimeReceived, err
	}, stmt, nil
}

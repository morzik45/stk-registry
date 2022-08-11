package postgres

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type CorrectPersonData struct {
	Snils      string    `db:"snils"`
	Family     string    `db:"family"`
	Name       string    `db:"name"`
	Patronymic string    `db:"patronymic"`
	Birthdate  time.Time `db:"birthdate"`
}

type CorrectPersonsData struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	searchSnils   func(ctx context.Context, person *CorrectPersonData) error
	searchBySnils func(ctx context.Context, person *CorrectPersonData) error
}

func NewCorrectPersonsData(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*CorrectPersonsData, error) {
	cpd := CorrectPersonsData{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := cpd.initCorrectPersonsData(ctxShort)
	if err != nil {
		logger.Error("failed to init correctPersonsData", zap.Error(err))
		return nil, err
	}
	return &cpd, nil
}

func (cpd *CorrectPersonsData) Close() error {
	for _, stmt := range cpd.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cpd *CorrectPersonsData) initCorrectPersonsData(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	cpd.searchSnils, stmt, err = cpd.initSearchSnils(ctx)
	if err != nil {
		return
	}
	cpd.stmts = append(cpd.stmts, stmt)

	cpd.searchBySnils, stmt, err = cpd.initSearchBySnils(ctx)
	if err != nil {
		return
	}
	cpd.stmts = append(cpd.stmts, stmt)

	return
}

func (cpd *CorrectPersonsData) SearchSnils(ctx context.Context, person *CorrectPersonData) (err error) {
	if cpd.searchSnils == nil {
		return errors.New("searchSnils func is not initialized")
	}
	return cpd.searchSnils(ctx, person)
}

func (cpd *CorrectPersonsData) initSearchSnils(ctx context.Context) (func(ctx context.Context, person *CorrectPersonData) error, *sqlx.NamedStmt, error) {
	query := `
		SELECT "snils", "family", "name", "patronymic", "birthdate"
		FROM correct_person_data
		WHERE "family" = :family 
		  AND "name" = :name 
		  AND "patronymic" = :patronymic 
		  AND "birthdate" = :birthdate
	`
	stmt, err := cpd.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, person *CorrectPersonData) (err error) {
		return stmt.GetContext(ctx, person, person)
	}, stmt, nil
}

func (cpd *CorrectPersonsData) SearchBySnils(ctx context.Context, person *CorrectPersonData) (err error) {
	if cpd.searchBySnils == nil {
		return errors.New("searchBySnils func is not initialized")
	}
	return cpd.searchBySnils(ctx, person)
}

func (cpd *CorrectPersonsData) initSearchBySnils(ctx context.Context) (func(ctx context.Context, person *CorrectPersonData) error, *sqlx.NamedStmt, error) {
	query := `
		SELECT "snils", "family", "name", "patronymic", "birthdate"
		FROM correct_person_data
		WHERE "snils" = :snils
	`
	stmt, err := cpd.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, person *CorrectPersonData) (err error) {
		return stmt.GetContext(ctx, person, person)
	}, stmt, nil
}

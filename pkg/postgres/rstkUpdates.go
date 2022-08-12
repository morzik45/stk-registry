package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type RstkUpdate struct {
	ID         int       `db:"id"`
	UploadedAt time.Time `db:"uploaded_at"`
	TypeID     int       `db:"type_id"`
	FromDate   time.Time `db:"from_date"`
}

type RstkUpdateInfo struct {
	ID         int64           `db:"id" json:"id"`
	TypeID     int             `db:"type_id" json:"type_id"`
	UploadedAt time.Time       `db:"uploaded_at" json:"uploaded_at"`
	FromDate   time.Time       `db:"from_date" json:"from_date"`
	Lines      int             `db:"lines" json:"lines"`
	Errors     json.RawMessage `db:"errors" json:"errors"`
}

type RstkUpdateReportForERC struct {
	FullName string    `db:"full_name" json:"full_name"`
	Snils    string    `db:"snils" json:"snils"`
	Date     time.Time `db:"date" json:"date"`
}

type RstkUpdates struct {
	db     *sqlx.DB
	stmts  []*sqlx.NamedStmt
	logger *zap.Logger

	create               func(ctx context.Context, rstkUpdate *RstkUpdate, tx *sqlx.Tx) error
	getInfo              func(ctx context.Context) ([]RstkUpdateInfo, error)
	delete               func(ctx context.Context, id int64, tx *sqlx.Tx) error
	reportForERC         func(ctx context.Context, from, to int64) ([]RstkUpdateReportForERC, error)
	reportForErcWithMark func(ctx context.Context, tx *sqlx.Tx) ([]RstkUpdateReportForERC, error)
}

func NewRstkUpdates(ctx context.Context, db *sqlx.DB, logger *zap.Logger) (*RstkUpdates, error) {
	ru := RstkUpdates{
		db:     db,
		logger: logger,
	}
	ctxShort, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := ru.initRstkUpdates(ctxShort)
	if err != nil {
		logger.Error("failed to init rstkUpdates", zap.Error(err))
		return nil, err
	}
	return &ru, nil
}

func (ru *RstkUpdates) Close() error {
	for _, stmt := range ru.stmts {
		if stmt != nil {
			err := stmt.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ru *RstkUpdates) initRstkUpdates(ctx context.Context) (err error) {
	var stmt *sqlx.NamedStmt
	ru.create, stmt, err = ru.initCreate(ctx)
	if err != nil {
		return
	}
	ru.stmts = append(ru.stmts, stmt)

	ru.getInfo, stmt, err = ru.initGetInfo(ctx)
	if err != nil {
		return
	}
	ru.stmts = append(ru.stmts, stmt)

	ru.delete, stmt, err = ru.initDelete(ctx)
	if err != nil {
		return
	}
	ru.stmts = append(ru.stmts, stmt)

	ru.reportForERC, stmt, err = ru.initReportForERC(ctx)
	if err != nil {
		return
	}
	ru.stmts = append(ru.stmts, stmt)

	ru.reportForErcWithMark, stmt, err = ru.initReportForErcWithMark(ctx)
	if err != nil {
		return
	}
	ru.stmts = append(ru.stmts, stmt)

	return
}

// Create добавляет новый отчёт о выданных картах в БД
func (ru *RstkUpdates) Create(ctx context.Context, update *RstkUpdate, tx *sqlx.Tx) error {
	if ru.create == nil {
		return errors.New("create func is not defined")
	}
	return ru.create(ctx, update, tx)
}

func (ru *RstkUpdates) initCreate(ctx context.Context) (func(ctx context.Context, update *RstkUpdate, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	stmt, err := ru.db.PrepareNamedContext(ctx, `
		INSERT INTO rstk_updates ("type_id", "from_date")
		VALUES (:type_id, :from_date)
		RETURNING id`)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, update *RstkUpdate, tx *sqlx.Tx) error {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		return currentStmt.GetContext(ctx, &update.ID, *update)
	}, stmt, nil
}

// GetInfo собирает информацию о загруженных отчётах для отображения на сайте
func (ru *RstkUpdates) GetInfo(ctx context.Context) ([]RstkUpdateInfo, error) {
	if ru.getInfo == nil {
		return nil, errors.New("getInfo func is not defined")
	}
	return ru.getInfo(ctx)
}

func (ru *RstkUpdates) initGetInfo(ctx context.Context) (func(ctx context.Context) ([]RstkUpdateInfo, error), *sqlx.NamedStmt, error) {
	stmt, err := ru.db.PrepareNamedContext(ctx, `
		SELECT ru.id, ru."type_id", ru.uploaded_at, ru.from_date,
		       COALESCE((SELECT COUNT(*) FROM persons_from_rstk WHERE rstk_update_id = ru.id), 0) AS lines,
		       	COALESCE((SELECT to_json(array_agg(row_to_json(d)))
				FROM (SELECT "id",
							 pfr."snils"                                                  as "snils",
							 pfr."family" || ' ' || pfr."name" || ' ' || pfr."patronymic" AS "full_name",
							 pfr."errors"
					  FROM persons_from_rstk pfr
					  WHERE pfr."rstk_update_id" = ru."id" AND pfr."errors" IS NOT NULL) d), '[]')    AS "errors"
		FROM rstk_updates AS ru
		ORDER BY ru.uploaded_at DESC;
		`)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context) (rui []RstkUpdateInfo, err error) {
		err = stmt.SelectContext(ctx, &rui, map[string]interface{}{})
		return
	}, stmt, nil
}

// Delete удаляет отчёт о выданных картах, вместе с ним каскадно удаляет все записи из persons_from_rstk (например, чтобы исправить и добавить заново)
func (ru *RstkUpdates) Delete(ctx context.Context, id int64, tx *sqlx.Tx) error {
	if ru.delete == nil {
		return errors.New("delete func is not defined")
	}
	return ru.delete(ctx, id, tx)
}

func (ru *RstkUpdates) initDelete(ctx context.Context) (func(ctx context.Context, id int64, tx *sqlx.Tx) error, *sqlx.NamedStmt, error) {
	stmt, err := ru.db.PrepareNamedContext(ctx, `DELETE FROM rstk_updates WHERE id = :id`)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, id int64, tx *sqlx.Tx) (err error) {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		_, err = currentStmt.ExecContext(ctx, map[string]interface{}{"id": id})
		return
	}, stmt, nil
}

// ReportForERC собирает данные для отправки в ЕРЦ за указанный период (не помечая как отправленные, просто для теста/информации)
func (ru *RstkUpdates) ReportForERC(ctx context.Context, from, to int64) ([]RstkUpdateReportForERC, error) {
	if ru.reportForERC == nil {
		return nil, errors.New("reportForErcRange func is not defined")
	}
	return ru.reportForERC(ctx, from, to)
}

func (ru *RstkUpdates) initReportForERC(ctx context.Context) (func(ctx context.Context, from, to int64) ([]RstkUpdateReportForERC, error), *sqlx.NamedStmt, error) {
	stmt, err := ru.db.PrepareNamedContext(ctx, `
		SELECT pfr.snils,
			   pfr.family || ' ' || pfr.name || ' ' || pfr.patronymic AS full_name,
			   pfr."date"
		FROM persons_from_rstk pfr
				 LEFT JOIN persons_from_erc pe ON pe.snils = pfr.snils
				 LEFT JOIN sent_to_erc ste ON pfr.snils = ste.snils
		WHERE pe.snils IS NULL -- исключаем тех кто уже покупал талоны (его карта уже заблокирована, до конца текущего периода он продолжает пользовать талонами)
		  AND ste.snils IS NULL -- исключаем тех кого уже отправляли в ЕРЦ
		  AND (pfr."date" >= to_timestamp(:from) OR :from = 0)
		  AND (pfr."date" <= to_timestamp(:to) OR :to = 0);`,
	)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, from, to int64) (rui []RstkUpdateReportForERC, err error) {
		err = stmt.SelectContext(ctx, &rui, map[string]interface{}{
			"from": from,
			"to":   to,
		})
		return
	}, stmt, nil
}

// ReportForErcWithMark собирает данные для отправки в ЕРЦ и помечает как отправленные
func (ru *RstkUpdates) ReportForErcWithMark(ctx context.Context, tx *sqlx.Tx) ([]RstkUpdateReportForERC, error) {
	if ru.reportForErcWithMark == nil {
		return nil, errors.New("reportForErcRange func is not defined")
	}
	return ru.reportForErcWithMark(ctx, tx)
}

func (ru *RstkUpdates) initReportForErcWithMark(ctx context.Context) (func(ctx context.Context, tx *sqlx.Tx) ([]RstkUpdateReportForERC, error), *sqlx.NamedStmt, error) {
	stmt, err := ru.db.PrepareNamedContext(ctx, `
		WITH a AS (SELECT pfr.snils,
						  pfr.family || ' ' || pfr.name || ' ' || pfr.patronymic AS full_name,
						  pfr."date"
				   FROM persons_from_rstk pfr
							LEFT JOIN persons_from_erc pe ON pe.snils = pfr.snils
							LEFT JOIN sent_to_erc ste ON pfr.snils = ste.snils
				   WHERE pe.snils IS NULL -- исключаем тех кто уже покупал талоны (его карта уже заблокирована, до конца текущего периода он продолжает пользовать талонами)
					 AND ste.snils IS NULL -- исключаем тех кого уже отправляли в ЕРЦ 
		), b AS (INSERT INTO sent_to_erc (snils, "date") -- помечаем как отправленные
				   SELECT snils, CURRENT_TIMESTAMP
				   FROM a
		) SELECT "snils", "full_name", "date" FROM a;`,
	)
	if err != nil {
		return nil, nil, err
	}
	return func(ctx context.Context, tx *sqlx.Tx) (rui []RstkUpdateReportForERC, err error) {
		currentStmt := stmt
		if tx != nil {
			currentStmt = tx.NamedStmtContext(ctx, stmt)
		}
		err = currentStmt.SelectContext(ctx, &rui, map[string]interface{}{})
		return
	}, stmt, nil
}

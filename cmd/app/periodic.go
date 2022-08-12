package main

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/morzik45/stk-registry/pkg/email/sender"
	"github.com/morzik45/stk-registry/pkg/scheduler"
	"github.com/morzik45/stk-registry/pkg/utils"
	"go.uber.org/zap"
	"io"
	"time"
)

func (app *App) RunPeriodicTasks() {

	// Периодически проверяем почту на новые сообщения от ЕРЦ.
	// Откладываем проверку на минуту для ожидания полной инициализации приложения и
	// спама при падении приложения после запуска и постоянного рестарта.
	if len(app.cfg.Email.FromErc) > 0 {
		app.emailCheckerScheduler = scheduler.NewTimedExecutor(
			time.Minute,
			app.cfg.Email.CheckInterval,
		)
		app.emailCheckerScheduler.Start(func() {
			defer utils.Recover(app.logger)
			isHaveNew, err := app.emailReceiver.GetNewFromErc()
			if err != nil {
				app.logger.Error("failed to get new from erc", zap.Error(err))
			}
			if isHaveNew {
				app.logger.Info("new erc message found")
			}
		}, true)
	}
	// Раз в сутки отправляем отчёт о выданных картах в ЕРЦ(если есть новые карты).
	// Рассчитываем время до ближайшей отправки отчёта о картах в ЕРЦ.
	startTime := time.Time(app.cfg.Email.SendReportAt)
	if startTime.Before(time.Now()) { // если время отправки отчёта уже прошло
		startTime = startTime.Add(time.Hour * 24) // то начинаем с завтрашнего дня
		app.logger.Info("Следующий отчёт о картах будет отправлен", zap.Time("at", startTime))
	}
	app.emailSenderScheduler = scheduler.NewTimedExecutor(
		time.Until(startTime), // сколько времени до первого запуска
		time.Hour*24,          // периодичность запуска
	)
	app.emailSenderScheduler.Start(func() {
		defer utils.Recover(app.logger)
		ctxMinute, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		err := app.MakeAndSendReportToERC(ctxMinute)
		if err != nil {
			app.logger.Error("failed to make and send report to erc", zap.Error(err))
		}
	}, true)
}

// MakeAndSendReportToERC отправляет отчёт в ЕРЦ по выбранным картам
func (app *App) MakeAndSendReportToERC(ctx context.Context) error {
	// Открываем транзакцию, если не получится отправить в ерц, то откатиться и не помечать как отправленное
	tx, err := app.db.BeginTx(ctx)
	if err != nil {
		app.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer func(tx *sqlx.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Собираем данные для отчета
	r, err := app.db.RstkUpdates.ReportForErcWithMark(ctx, tx)

	// Если нет данных для отчета, заканчиваем работу
	if len(r) == 0 {
		return nil
	}

	// Создаем отчет в формате xlsx
	buf, err := utils.MakeReportForErc(r)
	if err != nil {
		app.logger.Error("failed to make report", zap.Error(err))
		return err
	}

	// Отправляем отчет в ерц
	err = sender.SendFiles(
		[]io.Reader{buf},
		app.cfg.Email.ToErc,
		fmt.Sprintf("МКУ ТУ Реестр выданых карт за %s", time.Now().Format("02.01.2006")),
		app.cfg,
	)
	if err != nil {
		app.logger.Error("failed to send report", zap.Error(err))
		return err
	}

	// Commit транзакцию
	err = tx.Commit()
	if err != nil {
		app.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}
	return nil
}

package receiver

import (
	"bytes"
	"context"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-pop3"
	"github.com/morzik45/stk-registry/pkg/config"
	"github.com/morzik45/stk-registry/pkg/persons"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

// TODO: Это всё надо нещадно рефакторить, накидывал на скорость.

type Receiver struct {
	client    *pop3.Client
	conn      *pop3.Conn
	connMutex sync.Mutex
	logger    *zap.Logger
	config    *config.Config
	db        *postgres.DB
}

func NewReceiver(db *postgres.DB, cfg *config.Config, logger *zap.Logger) (*Receiver, error) {
	// Initialize the client.
	p := pop3.New(pop3.Opt{
		Host:       cfg.Email.Host,
		Port:       cfg.Email.PortPOP3,
		TLSEnabled: false,
	})

	r := Receiver{
		client: p,
		config: cfg,
		db:     db,
		logger: logger.Named("email_receiver"),
	}

	return &r, nil
}

func (r *Receiver) GetLastEmail() (last time.Time) {
	var err error
	last, err = r.db.Emails.GetLastReceivedTime(context.TODO(), nil)
	if err != nil {
		r.logger.Error("Error getting last email", zap.Error(err))
		last = time.Time(r.config.InitDate)
	}
	return
}

// Connect to the server.
func (r *Receiver) connect() (err error) {
	r.connMutex.Lock()
	var isConnected bool
	defer func(isConnected *bool) {
		// Отпускаем блокировку потому что не получилось подключиться
		// и метод disconnect не вызовется, блокировка останется навечно.
		if !*isConnected {
			r.connMutex.Unlock()
		}
	}(&isConnected)

	r.logger.Info("Connecting to mail server...")
	r.conn, err = r.client.NewConn()
	if err != nil {
		r.logger.Error("Error connecting to mail server",
			zap.String("host", r.config.Email.Host),
			zap.Int("port", r.config.Email.PortPOP3),
			zap.Error(err),
		)
		return err
	}
	if err = r.conn.Auth(r.config.Email.Username, r.config.Email.Password); err != nil {
		r.logger.Error("Error authenticating to mail server",
			zap.String("username", r.config.Email.Username),
			zap.Error(err),
		)
		return err
	}
	isConnected = true
	r.logger.Info("Connected to mail server.")
	return nil
}

// Disconnect from the server.
func (r *Receiver) disconnect() {
	r.logger.Info("Disconnecting from mail server...")
	err := r.conn.Quit()
	if err != nil {
		r.logger.Error("Error disconnecting from mail server",
			zap.Error(err),
		)
	}
	r.conn = nil
	r.logger.Info("Disconnected from mail server.")
	r.connMutex.Unlock()
}

func (r *Receiver) GetNewFromErc() (isHaveNew bool, err error) {
	return r.Receive(r.config.Email.FromErc, r.GetLastEmail())
}

func (r *Receiver) Receive(from string, afterTime time.Time) (isHaveNew bool, err error) {
	if err = r.connect(); err != nil {
		return
	}
	defer r.disconnect()
	//Print the total number of messages and their size.
	count, _, err := r.conn.Stat()
	if err != nil {
		r.logger.Error("Ошибка получения количества сообщений", zap.Error(err))
		return
	}
	r.logger.Debug("Всего сообщений на сервере:", zap.Int("count", count))

	// Pull all messages on the server. Message IDs go from count to 1.
	for id := count; id > 0; id-- {
		// Получим тело сообщения
		var mes *bytes.Buffer
		mes, err = r.conn.RetrRaw(id)
		if err != nil {
			r.logger.Error("Ошибка при получении сообщения", zap.Int("id", id), zap.Error(err))
			continue
		}

		// Парсим сообщение
		isNeedMore, _isHaveNew := r.parseMessage(mes.Bytes(), from, afterTime)
		if !isHaveNew && _isHaveNew { // Если не переворачивали флаг ранее и есть новые данные
			isHaveNew = true
		}
		if !isNeedMore {
			// Если вернулось false, то пошли уже старые сообщения и необходимо остановиться
			break
		}

	}
	return
}

func (r *Receiver) getTypeID(from string) int {
	switch from {
	case r.config.Email.FromErc:
		return 1
	default:
		return 0
	}
}

func (r *Receiver) parseMessage(body []byte, from string, afterTime time.Time) (isNeedMore, isHaveNew bool) {
	var err error
	isNeedMore = true // по умолчанию нужно продолжать получать сообщения

	// Парсим сообщение
	var mr *mail.Reader
	mr, err = mail.CreateReader(bytes.NewReader(body))
	if err != nil {
		r.logger.Error("Error creating mail reader", zap.Error(err))
		return
	}

	// Получаем информацию о письме
	var e postgres.Email
	header := mr.Header
	if e.DatetimeReceived, err = header.Date(); err != nil {
		r.logger.Error("Ошибка получения даты", zap.Error(err))
		return
	}
	if !e.DatetimeReceived.After(afterTime) {
		r.logger.Info("Письмо старше чем последнее полученное", zap.Time("datetime", e.DatetimeReceived), zap.Time("last", afterTime))
		isNeedMore = false // не нужно продолжать получать сообщения
		return
	}

	var fromAddr []*mail.Address
	if fromAddr, err = header.AddressList("From"); err == nil {
		e.FromAddress = fromAddr[0].Address
		if e.FromAddress != from {
			r.logger.Info("Email from address is not expected", zap.String("from", e.FromAddress), zap.String("expected", from))
			// Мы ждём письмо от нужного адреса, но получили письмо от другого адреса, просто пропускаем
			return
		}
		e.TypeID = r.getTypeID(e.FromAddress)
	} else {
		r.logger.Error("Error getting from address", zap.Error(err))
		return
	}

	e.MessageID, err = header.MessageID()
	if err != nil {
		r.logger.Error("Error getting message id", zap.Error(err))
		return
	}
	e.DatetimeParsed = time.Now() // Время парсинга письма
	e.File = body                 // Сохраняем письмо в базу данных

	// На этом этапе мы получили всю информацию о письме. Сохраним ее в транзакции, чтобы получить ее идентификатор.
	// Создадим транзакцию для записи в БД
	var tx *sqlx.Tx
	tx, err = r.db.BeginTx(context.TODO())
	if err != nil {
		r.logger.Error("Error starting transaction", zap.Error(err))
		return
	}
	defer func(tx *sqlx.Tx) { _ = tx.Rollback() }(tx)

	err = r.db.Emails.Create(context.TODO(), &e, tx)
	if err != nil {
		r.logger.Error("Error creating email in db", zap.Error(err))
		return
	}

	// Ищем вложения в письме.
	for {
		var part *mail.Part
		part, err = mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			r.logger.Error("Error getting next part", zap.Error(err))
			break
		}
		switch h := part.Header.(type) {
		case *mail.AttachmentHeader:
			var eu postgres.ErcUpdate
			eu.EmailID = e.ID
			eu.Name, err = h.Filename()
			if err != nil {
				r.logger.Error("Error getting filename", zap.Error(err))
				continue
			}

			// Сохраняем вложение в транзакции.
			err = r.db.ErcUpdates.Create(context.TODO(), &eu, tx)
			if err != nil {
				r.logger.Error("Error creating erc update", zap.Error(err))
				continue
			}

			ps := persons.ParseDocumentFromErc(part.Body, r.db.CorrectPersonsData)
			if len(ps) == 0 {
				r.logger.Info("No persons found in attachment", zap.String("filename", eu.Name))
				continue
			} else {
				if !isHaveNew {
					isHaveNew = true // Есть новые данные
				}
			}
			for i := range ps {
				ps[i].ErcUpdateID = eu.ID
			}
			err = r.db.PersonsFromErc.CreateMany(context.TODO(), ps, tx)
			if err != nil {
				r.logger.Error("Error creating persons from erc", zap.Error(err))
				continue
			}

		}
	}

	// Закроем транзакцию сохранения в БД
	err = tx.Commit()
	if err != nil {
		r.logger.Error("Error committing transaction", zap.Error(err))
	}
	return
}

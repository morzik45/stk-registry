package sender

import (
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/morzik45/stk-registry/pkg/config"
	"io"
	"net/smtp"
	"time"
)

// Небольшой хак для авторизации на почте без SSL
type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

// SendFiles отправляет !Excel! файлы на почту
func SendFiles(readers []io.Reader, to []string, subject string, cfg *config.Config) error {

	// Подготовка письма
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", cfg.Organization, cfg.Email.Username) // От кого
	e.To = to                                                             // Кому
	e.Subject = subject                                                   // Тема

	// Прикрепляем !Excel! файлы
	for _, r := range readers {
		_, err := e.Attach(
			r,
			time.Now().Format("20060201150405")+".xlsx",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		)
		if err != nil {
			return err
		}
	}

	// Отправляем письмо
	return e.Send(
		fmt.Sprintf("%s:%d", cfg.Email.Host, cfg.Email.PortSMTP),
		unencryptedAuth{smtp.PlainAuth("", cfg.Email.Username, cfg.Email.Password, cfg.Email.Host)},
	)
}

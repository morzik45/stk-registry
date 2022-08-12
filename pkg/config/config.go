package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"sync"
	"time"
)

var config *Config
var once sync.Once

// TimeToday В конфиге задаётся время в формате "15:04:05" или "15:04", дата будет взята из текущего времени.
type TimeToday time.Time

func (t *TimeToday) UnmarshalText(text []byte) error {
	tt, err := time.Parse("15:04:05", string(text))
	if err != nil {
		tt, err = time.Parse("15:04", string(text))
	}
	td := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), tt.Hour(), tt.Minute(), tt.Second(), 0, time.Local)
	*t = TimeToday(td)
	return err
}

type Date time.Time

func (t *Date) UnmarshalText(text []byte) error {
	tt, err := time.Parse("2006-01-02", string(text))
	if err != nil {
		tt, err = time.Parse("02.01.2006", string(text))
	}
	*t = Date(tt)
	return err
}

type Config struct {
	Debug        bool   `env:"IS_DEV" envDefault:"true"`
	Organization string `env:"ORGANIZATION" envDefault:"ООО ТЕСТ"`
	InitDate     Date   `env:"INIT_DATE" envDefault:"2022-07-01"`
	Web          struct {
		LocalPort int    `env:"WEB_LOCAL_PORT" envDefault:"8080"`
		Hostname  string `env:"WEB_HOSTNAME"`
	}
	Email struct {
		Host          string        `env:"EMAIL_HOST"`
		PortPOP3      int           `env:"EMAIL_PORT_POP3" envDefault:"110"`
		PortSMTP      int           `env:"EMAIL_PORT_SMTP" envDefault:"25"`
		Username      string        `env:"EMAIL_USERNAME"`
		Password      string        `env:"EMAIL_PASSWORD"`
		FromErc       string        `env:"EMAIL_FROM_ERC"`
		ToErc         []string      `env:"EMAIL_TO_ERC"`
		SendReportAt  TimeToday     `env:"EMAIL_SEND_REPORT_AT" envDefault:"06:00"`
		CheckInterval time.Duration `env:"EMAIL_CHECK_INTERVAL" envDefault:"30m"`
	}
	Postgres struct {
		Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
		Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
		DBName   string `env:"POSTGRES_DB" envDefault:"postgres"`
		Username string `env:"POSTGRES_USER" envDefault:"postgres"`
		Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		SSLMode  string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
	}
	Logger struct {
		Level    string `env:"LOG_LEVEL" envDefault:"debug"`
		Path     string `env:"LOG_PATH" envDefault:"/var/log/vkdumps/log.log"`
		Telegram struct {
			Enabled  bool   `env:"LOG_TG_ENABLED" envDefault:"false"`
			Token    string `env:"LOG_TG_TOKEN"`
			UsersIDs []int  `env:"LOG_TG_USERS_IDS"`
			Level    string `env:"LOG_TG_LEVEL" envDefault:"warn"`
		}
	}
}

func GetConfig() *Config {
	var c Config
	if config == nil {
		once.Do(func() {
			if err := env.Parse(&c); err != nil {
				log.Fatalf("%+v\n", err)
			}
			config = &c
		})
	}
	log.Println("Load config:", config)
	return config
}

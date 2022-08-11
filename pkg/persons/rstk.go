package persons

import (
	"bufio"
	"fmt"
	"github.com/morzik45/stk-registry/pkg/parser"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/morzik45/stk-registry/pkg/utils"
	"io"
	"log"
	"strings"
)

func Trim(data string) string {
	return strings.TrimSpace(strings.TrimRight(strings.TrimLeft(data, "'"), "'"))
}

func ParseRowFromRSTK(data string) (postgres.PersonFromRSTK, error) {
	var r postgres.PersonFromRSTK
	var err error
	rows := strings.Split(data, ",")
	if len(rows) != 4 {
		return postgres.PersonFromRSTK{}, fmt.Errorf("invalid row: %s", data)
	}

	r.Snils, err = parser.Snils(Trim(rows[1]))
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	fio := strings.Split(Trim(rows[0]), " ")
	if len(fio) != 3 {
		return postgres.PersonFromRSTK{}, fmt.Errorf("invalid row: %s", data)
	}
	r.Family, err = parser.String(fio[0])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
	r.Name, err = parser.String(fio[1])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
	r.Patronymic, err = parser.String(fio[2])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Date, err = parser.Date(Trim(rows[2]))
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Number, err = parser.String(Trim(rows[3]))
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	return r, nil
}

func ParseDocumentFromRSTK(reader io.Reader) (rs []postgres.PersonFromRSTK, type_ int) {
	var err error

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var line string
		line, err = utils.StringFromWindows1251(scanner.Text())
		if err != nil {
			log.Println(err)
			continue
		}

		// Сработает только на первой строке и определит тип или выкинет ошибку
		// FIXME: Всегда будет только эти 2 типа? С точно такой формулировкой?
		if type_ == 0 {
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "список социальных карт":
				type_ = 1
			case "список банковских карт":
				type_ = 2
			default:
				log.Printf("invalid document, unknown type: %s", line)
				return nil, 0
			}
		}

		// Пропускаем пустые строки
		if len(line) < 2 {
			continue
		}

		var r postgres.PersonFromRSTK
		r, err = ParseRowFromRSTK(line)
		if err != nil {
			log.Println(err)
			continue
		}
		rs = append(rs, r)
	}
	return
}

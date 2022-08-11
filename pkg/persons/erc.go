package persons

import (
	"bufio"
	"context"
	"fmt"
	"github.com/morzik45/stk-registry/pkg/parser"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/morzik45/stk-registry/pkg/utils"
	"io"
	"log"
	"strings"
)

func parseRowFromErc(data string, correctData *postgres.CorrectPersonsData) (r postgres.PersonFromERC, err error) {
	rows := strings.Split(data, "|")
	if len(rows) != 13 {
		return postgres.PersonFromERC{}, fmt.Errorf("invalid row: %s", string(data))
	}

	// TODO: Переписать, полная хрень...
	var snilsErr, birthDateErr, familyErr, nameErr, PatronymicErr error
	r.Snils, snilsErr = parser.Snils(rows[0])
	r.Birthdate, birthDateErr = parser.Date(rows[1])
	r.Family, familyErr = parser.String(rows[2])
	r.Name, nameErr = parser.String(rows[3])
	r.Patronymic, PatronymicErr = parser.String(rows[4])

	if (birthDateErr != nil || familyErr != nil || nameErr != nil || PatronymicErr != nil) && snilsErr == nil {
		// если ошибка в дате или в фамилии, или в имени, или в отчестве, но нет ошибки в СНИЛСе, то ищем по СНИЛСу
		person := postgres.CorrectPersonData{
			Snils: r.Snils,
		}
		err = correctData.SearchBySnils(context.TODO(), &person)
		if err == nil {
			// если нашли, то заполняем поля по найденной записи
			r.Birthdate = person.Birthdate
			r.Family = person.Family
			r.Name = person.Name
			r.Patronymic = person.Patronymic
			// и обнуляем ошибки
			birthDateErr = nil
			familyErr = nil
			nameErr = nil
			PatronymicErr = nil
		}
	} else if snilsErr != nil && (birthDateErr == nil || familyErr == nil || nameErr == nil || PatronymicErr == nil) {
		// если ошибка в СНИЛСе и нет ошибок в дате, фамилии, имени и отчестве - ищем по ним
		person := postgres.CorrectPersonData{
			Birthdate:  r.Birthdate,
			Family:     r.Family,
			Name:       r.Name,
			Patronymic: r.Patronymic,
		}
		err = correctData.SearchSnils(context.TODO(), &person)
		if err == nil {
			r.Snils = person.Snils // заполняем СНИЛС по найденной записи
			snilsErr = nil         // обнуляем ошибку в СНИЛСе
		}
	}

	// если не удалось исправить ошибки, то сохраняем их
	if birthDateErr != nil {
		r.Errors = append(r.Errors, birthDateErr.Error())
	}
	if familyErr != nil {
		r.Errors = append(r.Errors, familyErr.Error())
	}
	if nameErr != nil {
		r.Errors = append(r.Errors, nameErr.Error())
	}
	if PatronymicErr != nil {
		r.Errors = append(r.Errors, PatronymicErr.Error())
	}
	if snilsErr != nil {
		r.Errors = append(r.Errors, snilsErr.Error())
	}

	r.Year, err = parser.Year(rows[5])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Semester, err = parser.Semester(rows[6])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Color, err = parser.String(rows[7])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Count, err = parser.Int(rows[8])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Spent, err = parser.Int(rows[9])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.Date, err = parser.Date(rows[10])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.CashierID, err = parser.Int(rows[11])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.CashierName, err = parser.String(rows[12])
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	return r, nil
}

func ParseDocumentFromErc(reader io.Reader, correctData *postgres.CorrectPersonsData) (result []postgres.PersonFromERC) {
	var err error

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var line string
		line, err = utils.StringFromWindows1251(scanner.Text())
		if err != nil {
			log.Printf("error: %s", err)
			continue
		}
		if len(line) == 0 {
			continue
		}
		var n postgres.PersonFromERC
		n, err = parseRowFromErc(line, correctData)
		if err != nil {
			log.Println(err)
			continue
		}
		result = append(result, n)
	}

	return result
}

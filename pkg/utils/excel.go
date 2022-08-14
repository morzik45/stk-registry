package utils

import (
	"bytes"
	"errors"
	"github.com/morzik45/stk-registry/pkg/parser"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"io"
	"strconv"
	"time"
)

// MakeReportForErc формирует отчет в формате Excel для ЕРЦ
func MakeReportForErc(r []postgres.RstkUpdateReportForERC) (buf *bytes.Buffer, err error) {
	// TODO: собирать возможные ошибки (откуда им тут взяться?)
	file := excelize.NewFile()
	file.NewSheet("aСТК.xlsx")
	file.DeleteSheet("Sheet1")
	file.SetActiveSheet(0)
	file.SetCellValue("aСТК.xlsx", "A1", "№ п/п")
	file.SetCellValue("aСТК.xlsx", "B1", "Фамилия Имя Отчество")
	file.SetCellValue("aСТК.xlsx", "C1", "СНИЛС")
	file.SetCellValue("aСТК.xlsx", "D1", "Дата готовности к выдаче")

	for i, v := range r {
		file.SetCellValue("aСТК.xlsx", "A"+strconv.Itoa(i+2), strconv.Itoa(i+1))
		file.SetCellValue("aСТК.xlsx", "B"+strconv.Itoa(i+2), v.FullName)
		file.SetCellValue("aСТК.xlsx", "C"+strconv.Itoa(i+2), v.Snils)
		file.SetCellValue("aСТК.xlsx", "D"+strconv.Itoa(i+2), v.Date.Format("02.01.2006"))
	}
	file.SetColWidth("aСТК.xlsx", "A", "A", 7)
	file.SetColWidth("aСТК.xlsx", "B", "B", 35)
	file.SetColWidth("aСТК.xlsx", "C", "C", 15)
	file.SetColWidth("aСТК.xlsx", "D", "D", 25)

	buf, err = file.WriteToBuffer()
	return
}

func MakeBreakersReport(r []postgres.BreakerView) (buf *bytes.Buffer, err error) {
	file := excelize.NewFile()
	sheetName := time.Now().Format("02.01.2006")
	file.NewSheet(sheetName)
	file.DeleteSheet("Sheet1")
	file.SetActiveSheet(0)
	file.SetCellValue(sheetName, "A1", "№ п/п")
	file.SetCellValue(sheetName, "B1", "Дата готовности к выдаче")
	file.SetCellValue(sheetName, "C1", "Фамилия Имя Отчество")
	file.SetCellValue(sheetName, "D1", "СНИЛС")
	file.SetCellValue(sheetName, "E1", "PAN")
	file.SetCellValue(sheetName, "F1", "Статус")

	file.SetColWidth(sheetName, "A", "A", 7)
	file.SetColWidth(sheetName, "B", "B", 25)
	file.SetColWidth(sheetName, "C", "C", 35)
	file.SetColWidth(sheetName, "D", "F", 15)

	for i, v := range r {
		file.SetCellValue(sheetName, "A"+strconv.Itoa(i+2), strconv.Itoa(i+1))
		file.SetCellValue(sheetName, "B"+strconv.Itoa(i+2), v.Date.Format("02.01.2006"))
		file.SetCellValue(sheetName, "C"+strconv.Itoa(i+2), v.Name)
		file.SetCellValue(sheetName, "D"+strconv.Itoa(i+2), v.Snils)
		file.SetCellValue(sheetName, "E"+strconv.Itoa(i+2), v.Pan)
		if v.Checked {
			file.SetCellValue(sheetName, "F"+strconv.Itoa(i+2), "Заблокирован")
		}
	}
	buf, err = file.WriteToBuffer()
	return
}

func MakeExcelForCorrection(r []postgres.PersonFromErcForCorrection) (buf *bytes.Buffer, err error) {
	sheetName := time.Now().Format("02.01.2006")
	file := excelize.NewFile()

	// добавляем новый лист
	file.NewSheet(sheetName)
	file.DeleteSheet("Sheet1")
	file.SetActiveSheet(0)

	// Заполняем заголовок таблицы
	file.SetCellValue(sheetName, "A1", "№ п/п")
	file.SetCellValue(sheetName, "B1", "Фамилия")
	file.SetCellValue(sheetName, "C1", "Имя")
	file.SetCellValue(sheetName, "D1", "Отчество")
	file.SetCellValue(sheetName, "E1", "Дата рождения")
	file.SetCellValue(sheetName, "F1", "СНИЛС")

	// Устанавливаем ширину колонок
	file.SetColWidth(sheetName, "A", "A", 7)
	file.SetColWidth(sheetName, "B", "F", 15)

	// Делаем первую строку (заголовок) выделенной
	style, err := file.NewStyle(&excelize.Style{
		Border: []excelize.Border{{
			Type:  "thick",
			Color: "#000000",
		}},
		Font: &excelize.Font{
			Bold: true,
		},
	})
	if err != nil {
		return nil, err
	}
	file.SetCellStyle(sheetName, "A1", "F1", style)

	// Заполняем таблицу данными
	for i, v := range r {
		file.SetCellInt(sheetName, "A"+strconv.Itoa(i+2), v.ID)
		file.SetCellStr(sheetName, "B"+strconv.Itoa(i+2), v.Family)
		file.SetCellStr(sheetName, "C"+strconv.Itoa(i+2), v.Name)
		file.SetCellStr(sheetName, "D"+strconv.Itoa(i+2), v.Patronymic)
		file.SetCellStr(sheetName, "E"+strconv.Itoa(i+2), v.Birthdate.Format("02.01.2006"))
		file.SetCellStr(sheetName, "F"+strconv.Itoa(i+2), v.Snils)
	}

	buf, err = file.WriteToBuffer()
	return
}

func ParseExcelForCorrection(buf io.Reader, logger *zap.Logger) (r []postgres.PersonFromErcForCorrection, err error) {
	logger = logger.With(zap.String("func", "ParseExcelForCorrection"))
	file, err := excelize.OpenReader(buf)
	if err != nil {
		logger.Error("Ошибка открытия файла", zap.Error(err))
		return
	}
	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		logger.Error("Ошибка получения строк", zap.Error(err))
		return
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		var (
			p   postgres.PersonFromErcForCorrection
			err error
		)
		p.ID, err = parser.Int(row[0])
		if err != nil {
			logger.Error("Неверный формат поля ID", zap.Error(err), zap.String("row", row[0]))
			continue
		}
		p.Family, err = parser.String(row[1])
		if err != nil {
			logger.Error("Неверный формат поля Фамилия", zap.Error(err), zap.String("row", row[1]))
			continue
		}
		p.Name, err = parser.String(row[2])
		if err != nil {
			logger.Error("Неверный формат поля Имя", zap.Error(err), zap.String("row", row[2]))
			continue
		}
		p.Patronymic, err = parser.String(row[3])
		if err != nil {
			logger.Error("Неверный формат поля Отчество", zap.Error(err), zap.String("row", row[3]))
			continue
		}
		p.Birthdate, err = parser.Date(row[4])
		if err != nil {
			logger.Error("Неверный формат поля Дата рождения", zap.Error(err), zap.String("row", row[4]))
			continue
		}
		p.Snils, err = parser.Snils(row[5])
		if err != nil {
			logger.Error("Неверный формат поля СНИЛС", zap.Error(err), zap.String("row", row[5]))
			continue
		}
		r = append(r, p)
	}
	if len(r) == 0 {
		logger.Error("Нет данных для обработки")
		err = errors.New("нет данных для обработки")
	}
	return
}

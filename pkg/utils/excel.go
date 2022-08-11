package utils

import (
	"bytes"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/xuri/excelize/v2"
	"strconv"
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

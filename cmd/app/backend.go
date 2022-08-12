package main

import (
	"github.com/gin-gonic/gin"
	"github.com/morzik45/stk-registry/pkg/persons"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/morzik45/stk-registry/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

func (app *App) initBackend() {
	api := app.router.Group("/api")

	api.GET("/health", app.health)

	api.GET("/retiree", app.retiree)
	api.GET("/breakers", app.breakersView)
	api.POST("/breakers/check", app.breakersSet)

	updates := api.Group("/updates")
	updates.GET("", app.getUpdatesInfo)
	updates.POST("/uploadERC", app.uploadERC)
	updates.POST("/uploadRSTK", app.uploadRSTK)
	updates.DELETE("/rstk/:id", app.deleteRSTK)
	updates.POST("/make-rstk-excel", app.makeRstkExcel)

}

func (app *App) makeRstkExcel(c *gin.Context) {
	var dates []string
	var err error
	err = c.BindJSON(&dates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(dates) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат дат"})
		return
	}
	var from, to time.Time
	from, err = time.Parse("2006-01-02", dates[0])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат дат"})
		return
	}
	to, err = time.Parse("2006-01-02", dates[1])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат дат"})
		return
	}

	r, err := app.db.RstkUpdates.ReportForERC(c.Request.Context(), from.Unix(), to.Unix())

	buf, err := utils.MakeReportForErc(r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//c.Writer.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	//c.Writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	//c.Writer.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	//c.Writer.Write(buf.Bytes())
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (app *App) deleteRSTK(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	err = app.db.RstkUpdates.Delete(c.Request.Context(), id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (app *App) uploadRSTK(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fromDate time.Time

	if len(file.Filename) > 13 {
		fromDateStr := file.Filename[:10]
		fromDate, err = time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			fromDate, err = time.Parse("02.01.2006", fromDateStr)
			if err != nil {
				fromDate = time.Now()
			}
		}
	}

	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	p, t := persons.ParseDocumentFromRSTK(reader)
	if t == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось определить тип документа"})
		return
	}

	tx, err := app.db.BeginTx(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	ru := postgres.RstkUpdate{TypeID: t, FromDate: fromDate}
	err = app.db.RstkUpdates.Create(c.Request.Context(), &ru, tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range p {
		p[i].RstkUpdateID = ru.ID
	}
	err = app.db.PersonsFromRSTK.CreateMany(c.Request.Context(), p, tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (app *App) uploadERC(c *gin.Context) {
	_, err := app.emailReceiver.GetNewFromErc()
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (app *App) getUpdatesInfo(c *gin.Context) {
	erc, err := app.db.ErcUpdates.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	stat, err := app.db.ErcUpdates.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	errorsData, err := app.db.ErcUpdates.GetErrors(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	rstkUpdates, err := app.db.RstkUpdates.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status":      "ok",
		"erc":         erc,
		"stat":        stat,
		"errors_data": errorsData,
		"rstk":        rstkUpdates,
	})
}

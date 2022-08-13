package main

import (
	"github.com/gin-gonic/gin"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/morzik45/stk-registry/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

func (app *App) breakersSet(c *gin.Context) {
	var (
		breaker    postgres.Breaker
		ok         bool
		checkedStr string
		err        error
	)

	breaker.Snils, ok = c.GetQuery("snils")
	if !ok || breaker.Snils == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Не указан СНИЛС",
		})
		return
	}

	checkedStr, ok = c.GetQuery("checked")
	if !ok || checkedStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Не указано значение поля checked",
		})
		return
	}

	breaker.Checked, err = strconv.ParseBool(checkedStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Не верно указано значение поля checked",
		})
		return
	}

	err = app.db.Breakers.Create(c.Request.Context(), &breaker, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "ok",
		"data":   breaker,
	})
}

func (app *App) breakersView(c *gin.Context) {
	view, err := app.db.Breakers.GetView(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "ok",
		"data":   view,
	})
}

func (app *App) makeBreakersExcel(c *gin.Context) {
	var breakers []postgres.BreakerView
	var err error
	err = c.ShouldBindJSON(&breakers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(breakers) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны данные для экспорта"})
		return
	}

	buf, err := utils.MakeBreakersReport(breakers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Disposition", "attachment; filename=Нарушители_"+time.Now().Format("2006-01-02")+".xlsx")
	//c.Writer.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

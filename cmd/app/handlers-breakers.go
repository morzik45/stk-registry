package main

import (
	"github.com/gin-gonic/gin"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"net/http"
	"strconv"
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

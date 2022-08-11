package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func (app *App) retiree(c *gin.Context) {
	search := c.Query("search")
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
	offset, _ := strconv.ParseInt(c.Query("offset"), 10, 64)
	r, err := app.db.PersonsFromErc.Get(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(200, r)
}

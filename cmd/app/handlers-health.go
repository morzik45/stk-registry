package main

import "github.com/gin-gonic/gin"

func (app *App) health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

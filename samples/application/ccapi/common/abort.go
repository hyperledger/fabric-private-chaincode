package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Abort(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{
		"status": status,
		"error":  err.Error(),
	})
	c.Error(err)
}

func Respond(c *gin.Context, res interface{}, status int, err error) {
	if err != nil {
		c.JSON(status, gin.H{
			"response": res,
			"status":   status,
			"error":    err.Error(),
		})
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}

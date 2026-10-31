/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Abort responds with the given status and error
func Abort(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{
		"status": status,
		"error":  err.Error(),
	})
	c.Error(err)
}

// Respond responds with res, or with the given status and error if err is set
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

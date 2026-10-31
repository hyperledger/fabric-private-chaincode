/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-private-chaincode/samples/application/ccapi/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// AddRoutesToEngine registers the routes and handlers of the API
func AddRoutesToEngine(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/api-docs/index.html")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// serve swagger files
	docs.SwaggerInfo.BasePath = "/api"
	r.StaticFile("/swagger.yaml", "./docs/swagger.yaml")

	url := ginSwagger.URL("/swagger.yaml")
	r.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))

	addCCRoutes(r.Group("/api"))
}

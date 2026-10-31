package routes

import (
	"github.com/hyperledger/fabric-private-chaincode/samples/application/ccapi/handlers"

	"github.com/gin-gonic/gin"
)

func addCCRoutes(rg *gin.RouterGroup) {
	rg.POST("/:channelName/:chaincodeName/invoke/:txname", handlers.InvokeFpc)
	rg.PUT("/:channelName/:chaincodeName/invoke/:txname", handlers.InvokeFpc)
	rg.DELETE("/:channelName/:chaincodeName/invoke/:txname", handlers.InvokeFpc)
	rg.POST("/:channelName/:chaincodeName/query/:txname", handlers.QueryFpc)
	rg.GET("/:channelName/:chaincodeName/query/:txname", handlers.QueryFpc)

	rg.POST("/invoke/:txname/", handlers.InvokeFpcDefault)
	rg.POST("/invoke/:txname", handlers.InvokeFpcDefault)
	rg.PUT("/invoke/:txname/", handlers.InvokeFpcDefault)
	rg.PUT("/invoke/:txname", handlers.InvokeFpcDefault)
	rg.DELETE("/invoke/:txname/", handlers.InvokeFpcDefault)
	rg.DELETE("/invoke/:txname", handlers.InvokeFpcDefault)
	rg.POST("/query/:txname/", handlers.QueryFpcDefault)
	rg.POST("/query/:txname", handlers.QueryFpcDefault)
	rg.GET("/query/:txname/", handlers.QueryFpcDefault)
	rg.GET("/query/:txname", handlers.QueryFpcDefault)
}

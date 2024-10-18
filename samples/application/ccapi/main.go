package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/ccapi/chaincode"
	"github.com/hyperledger-labs/ccapi/server"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Create gin handler and start server
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:8080", // Test addresses
			"*",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Origin", "Content-Type"},
		AllowCredentials: true,
	}))
	go server.Serve(r, ctx)
	// Events are not integrated with FPC
	if os.Getenv("FPC_ENABLED") != "true" {

		// Register to chaincode events
		go chaincode.WaitForEvent(os.Getenv("CHANNEL"), os.Getenv("CCNAME"), "eventName", func(ccEvent *fab.CCEvent) {
			log.Println("Received CC event: ", ccEvent)
		})

		chaincode.RegisterForEvents()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	cancel()
}

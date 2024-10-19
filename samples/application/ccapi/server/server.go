package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger-labs/ccapi/common"
	"github.com/hyperledger-labs/ccapi/routes"
)

func defaultServer(r *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    ":80",
		Handler: r,
	}
}

// Serve starts the server with gin's default engine.
// Server gracefully shut's down
func Serve(r *gin.Engine, ctx context.Context) {
	// Defer close sdk to clear cache and free memory
	defer common.CloseSDK()

	if os.Getenv("FPC_ENABLED") == "true" {
		common.InitFpcConfig()
	}

	// Register routes and handlers
	routes.AddRoutesToEngine(r)

	// Returns a http.Server from provided handler
	srv := defaultServer(r)

	// listen and serve on 0.0.0.0:80 (for windows "localhost:80")
	go func(server *http.Server) {
		log.Println("Listening on port 80")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
	}(srv)

	// Graceful shutdown
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
	log.Println("Shutting down")
}

// Serve sync starts the server with a given wait group.
// When server starts, the wait group counter is increased and processes
// that depend on server can be ran synchronously with it
func ServeSync(ctx context.Context, wg *sync.WaitGroup) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	routes.AddRoutesToEngine(r)

	srv := defaultServer(r)

	go func(server *http.Server) {
		log.Println("Listening on port 80")
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Panic(err)
		}
		// finish wait group
		time.Sleep(1 * time.Second)
		wg.Done()
	}(srv)

	wg.Add(1)
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
	log.Println("Shutting down")
}

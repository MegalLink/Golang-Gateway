package main

import (
	"context"
	"fmt"
	"log"
	"megalink/gateway/client/channels"
	"megalink/gateway/client/connection"
	"megalink/gateway/client/handler"
	"megalink/gateway/client/heartbeat"
	"megalink/gateway/client/listener"
	"megalink/gateway/client/service"
	"megalink/gateway/client/sign"
	"megalink/gateway/client/types"
	"megalink/gateway/logger"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic occurred in main: ", err)
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()
	channel := channels.ProvideChannels[*types.ServerResponse]()

	ctx := context.Background()

	my_logger, err := logger.NewFastLogger()
	if err != nil {
		println("Error")
	}

	my_logger.Info("Main", "")
	id := uuid.New()
	fmt.Println(id.String())
	// Define server address for health check and heartbeat
	envVars := types.EnvVars{
		GinServerAdress:              "localhost:8080",
		FranchiseConnectionAdress:    "localhost:9090",
		ShowEcho:                     false,
		HeartSendBeatIntervalSeconds: 5,
		HeartBeatResponseWaitSeconds: 6,
	}

	signService := sign.NewSignService(&envVars)
	connFact := connection.NewConnFactory(&envVars)
	heartbeat := heartbeat.NewHeartBeatService(&envVars)
	connManager := connection.NewConnManager(signService, heartbeat, connFact, &envVars)
	errHandler := handler.NewErrorHandler()
	respHandler := handler.NewResponseHandler(ctx, channel)

	dataFastHandler := new(listener.ListenerChain).
		AddHandler(errHandler.HandleMessageError).
		AddHandler(heartbeat.HandleHeartBeatResponse).
		AddHandler(respHandler.HandleMessageResponse).
		BuildChain()

	// Listen DATAFAST response.
	listener := listener.NewListener(connManager, dataFastHandler, errHandler, &envVars)
	_ = connManager.SetupConnection(ctx)

	// ctx must be a context.Background() to listen forever.
	go listener.Listen(ctx)

	// Create a Gin router
	router := gin.New()

	router.Use(LoggingMiddleware(my_logger))
	router.Use(CustomRecoveryMiddleware(channel))

	sv := service.Service{
		Connection: connManager,
		Logger:     my_logger,
		Channel:    channel,
	}
	// Health check endpoint
	router.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})
	router.GET("/transaction", sv.TransactionService)

	srv := &http.Server{
		Addr:    envVars.GinServerAdress,
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		channel.CloseChannels()
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func CustomRecoveryMiddleware(channel *channels.ChannelStruct[*types.ServerResponse]) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				log.Printf("Panic recovered: %s\n", r)
				log.Printf("Stack trace: %s\n", debug.Stack())
				channel.CloseChannels()
				// Return a custom error response with the external message
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

func LoggingMiddleware(logger logger.IFastLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New()
		logger.WithPrefix(id.String())
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log details
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Error("ERROR", e)
			}
		} else {
			logger.Info("Loggin Middleware", fmt.Sprintf("%s %s %d %s %s\n", clientIP, method, path, statusCode, duration))
		}
	}
}

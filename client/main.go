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
	"runtime/debug"
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

	defer func() {
		println("Closing channels")
		channel.CloseChannels()
	}()

	global := "My global constant"
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
	router.Use(CustomRecoveryMiddleware(global))

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

	router.Run(envVars.GinServerAdress)
}

func CustomRecoveryMiddleware(message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				log.Printf("Panic recovered: %s\n", r)
				log.Printf("Stack trace: %s\n", debug.Stack())

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

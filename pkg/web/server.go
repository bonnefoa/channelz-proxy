package web

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func setupRouter(logger *zap.Logger) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware())
	skipLogs := []string{
		"/health",
	}
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, skipLogs...))
	router.Use(gin.Recovery())
	router.Use(gintrace.Middleware("channelz-proxy"))

	c := NewChannelzProxyRoutes(logger)
	router.GET("/readiness", c.readinessRoute)
	api := router.Group("/api")
	{
		api.GET("/channel", c.channelRoute)
		api.GET("/subchannel", c.subchannelRoute)
		api.GET("/socket", c.socketRoute)
		api.GET("/channels", c.channelsRoute)
		api.GET("/servers", c.serversRoute)
		api.GET("/serverSockets", c.serverSocketsRoute)
	}
	return router
}

func StartServer(ctx context.Context, addr string, logger *zap.Logger) {
	router := setupRouter(logger)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("error listen", zap.Error(err))
		}
	}()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", zap.Error(err))
	}
	logger.Info("Server exiting")
}

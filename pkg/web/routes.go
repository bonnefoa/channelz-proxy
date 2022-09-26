package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bonnefoa/channelz/channelz-proxy/pkg/grpc"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type ChannelzProxyRoutes struct {
	c      *grpc.ChannelzProxyServer
	logger *zap.Logger
}

func NewChannelzProxyRoutes(logger *zap.Logger) *ChannelzProxyRoutes {
	return &ChannelzProxyRoutes{
		c:      grpc.NewChannelzProxyServer(logger),
		logger: logger,
	}
}

func (s *ChannelzProxyRoutes) getHost(c *gin.Context) (string, error) {
	host, hasHost := c.GetQuery("host")
	if !hasHost {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing host parameter"})
		return "", errors.New("Missing host parameter")
	}
	return host, nil
}

func (s *ChannelzProxyRoutes) readinessRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Ok"})
}

func (s *ChannelzProxyRoutes) channelRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	channelId, err := strconv.Atoi(c.DefaultQuery("channelId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("channelId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channel, err := s.c.GetChannel(ctx, host, int64(channelId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting channel: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channel})
}

func (s *ChannelzProxyRoutes) subchannelRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	channelId, err := strconv.Atoi(c.DefaultQuery("subchannelId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("channelId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	subchannel, err := s.c.GetSubchannel(ctx, host, int64(channelId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting subchannel: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": subchannel})
}

func (s *ChannelzProxyRoutes) channelsRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	startId, err := strconv.Atoi(c.DefaultQuery("startId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("startId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channels, err := s.c.GetTopChannels(ctx, host, int64(startId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting channels: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels})
}

func (s *ChannelzProxyRoutes) socketRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	socketId, err := strconv.Atoi(c.DefaultQuery("socketId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("socketId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channels, err := s.c.GetSocket(ctx, host, int64(socketId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting channels: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels})
}

func (s *ChannelzProxyRoutes) serversRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	startId, err := strconv.Atoi(c.DefaultQuery("startId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("startId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	servers, err := s.c.GetServers(ctx, host, int64(startId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting servers: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": servers})
}

func (s *ChannelzProxyRoutes) serverSocketsRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	serverId, err := strconv.Atoi(c.DefaultQuery("serverId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("serverId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	startSocketId, err := strconv.Atoi(c.DefaultQuery("startSocketId", "0"))
	if err != nil {
		errMsg := fmt.Sprintf("startSocketId should be an int: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	sockets, err := s.c.GetServerSockets(ctx, host, int64(serverId), int64(startSocketId))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting server sockets: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sockets})
}

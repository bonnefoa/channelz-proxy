package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bonnefoa/channelz/channelz-proxy/pkg/grpc"
	"github.com/bonnefoa/channelz/channelz-proxy/pkg/util"
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing host parameter"})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "channelId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channel, err := s.c.GetChannel(ctx, host, int64(channelId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.FormatGrpcError(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channel})
}

// Get states of all subchannels of a channel
func (s *ChannelzProxyRoutes) channelSubchannelsRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	channelId, err := strconv.Atoi(c.DefaultQuery("channelId", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "channelId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	channel, err := s.c.GetChannel(ctx, host, int64(channelId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.FormatGrpcError(err))
		return
	}

	subchannelIds := make([]int64, 0)
	for _, subchannelRef := range channel.SubchannelRef {
		subchannelIds = append(subchannelIds, subchannelRef.SubchannelId)
	}

	subchannels, err := s.c.GetSubchannels(ctx, host, subchannelIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.FormatGrpcError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subchannels})
}

func (s *ChannelzProxyRoutes) subchannelsRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	subchannelIdsQuery := c.Query("subchannelIds")
	if subchannelIdsQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing subchannelIds parameter"})
		return
	}
	subchannelIdsStr := strings.Split(subchannelIdsQuery, ",")
	subchannelIds := make([]int64, 0)
	for _, subchannelIdStr := range subchannelIdsStr {
		subchannelId, err := strconv.Atoi(subchannelIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "subchannelId should be an int",
				"details": err.Error(),
			})
			return
		}
		subchannelIds = append(subchannelIds, int64(subchannelId))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	subchannels, err := s.c.GetSubchannels(ctx, host, subchannelIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.FormatGrpcError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subchannels})
}

func (s *ChannelzProxyRoutes) subchannelRoute(c *gin.Context) {
	host, err := s.getHost(c)
	if err != nil {
		return
	}
	channelId, err := strconv.Atoi(c.DefaultQuery("subchannelId", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "channelId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	subchannel, err := s.c.GetSubchannel(ctx, host, int64(channelId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting subchannel",
			"details": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "startId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channels, err := s.c.GetTopChannels(ctx, host, int64(startId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.FormatGrpcError(err))
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "socketId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	channels, err := s.c.GetSocket(ctx, host, int64(socketId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting channels",
			"details": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "startId should be an int",
			"details": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	servers, err := s.c.GetServers(ctx, host, int64(startId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting servers",
			"details": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "serverId should be an int",
			"details": err.Error()})
		return
	}

	startSocketId, err := strconv.Atoi(c.DefaultQuery("startSocketId", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "startSocketId should be an int",
			"details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	sockets, err := s.c.GetServerSockets(ctx, host, int64(serverId), int64(startSocketId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting server sockets",
			"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sockets})
}

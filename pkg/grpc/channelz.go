package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	channelzgrpc "google.golang.org/grpc/channelz/grpc_channelz_v1"
	"google.golang.org/grpc/credentials/insecure"
)

type ChannelzProxyServer struct {
	logger *zap.Logger

	cachedConn map[string]*grpc.ClientConn
}

func NewChannelzProxyServer(logger *zap.Logger) *ChannelzProxyServer {
	return &ChannelzProxyServer{
		logger:     logger.Named("ChannelzProxyServer"),
		cachedConn: make(map[string]*grpc.ClientConn),
	}
}

func (c *ChannelzProxyServer) getChannelClient(address string) (channelzgrpc.ChannelzClient, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	c.logger.Info("Connecting to grpc", zap.String("address", address))
	conn, ok := c.cachedConn[address]
	if ok {
		return channelzgrpc.NewChannelzClient(conn), nil
	}
	conn, err := grpc.Dial(address, dialOptions...)
	if err != nil {
		c.logger.Warn("Error dialing", zap.String("address", address), zap.Error(err))
		return nil, err
	}
	c.cachedConn[address] = conn
	client := channelzgrpc.NewChannelzClient(conn)
	return client, nil
}

func (c *ChannelzProxyServer) GetServers(ctx context.Context, address string, startServerId int64) ([]*channelzgrpc.Server, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	req := &channelzgrpc.GetServersRequest{StartServerId: startServerId}
	resp, err := clt.GetServers(ctx, req)
	if err != nil {
		c.logger.Warn("Error getting servers", zap.Error(err))
		return nil, err
	}
	return resp.Server, nil
}

func (c *ChannelzProxyServer) getServerSocketIds(ctx context.Context, clt channelzgrpc.ChannelzClient, serverId int64, startSocketId int64) ([]int64, error) {
	socketIds := make([]int64, 0)
	currentStart := startSocketId
	for {
		serverSocketReq := &channelzgrpc.GetServerSocketsRequest{ServerId: serverId, StartSocketId: currentStart}
		serverSocketResp, err := clt.GetServerSockets(ctx, serverSocketReq)
		if err != nil {
			c.logger.Warn("Error getting server sockets", zap.Error(err))
			return nil, err
		}
		nextSocketId := int64(0)
		for _, socketRef := range serverSocketResp.SocketRef {
			socketId := socketRef.GetSocketId()
			socketIds = append(socketIds, socketId)
			if socketId > nextSocketId {
				nextSocketId = socketId
			}
		}
		if nextSocketId == 0 || nextSocketId < startSocketId || serverSocketResp.End {
			return socketIds, nil
		}
		startSocketId = nextSocketId
	}
}

func (c *ChannelzProxyServer) GetServerSockets(ctx context.Context, address string, serverId int64, startSocketId int64) ([]*channelzgrpc.Socket, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	socketIds, err := c.getServerSocketIds(ctx, clt, serverId, startSocketId)
	if err != nil {
		return nil, err
	}
	sockets := make([]*channelzgrpc.Socket, 0)
	for _, socketId := range socketIds {
		c.logger.Info("Requesting socketId", zap.Int64("socketId", socketId))
		socketReq := &channelzgrpc.GetSocketRequest{SocketId: socketId}
		socketResp, err := clt.GetSocket(ctx, socketReq)
		if err != nil {
			c.logger.Warn("Error getting socket", zap.Error(err))
			return nil, err
		}
		sockets = append(sockets, socketResp.Socket)
	}
	return sockets, nil
}

func (c *ChannelzProxyServer) GetSocket(ctx context.Context, address string, socketId int64) (*channelzgrpc.Socket, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	req := &channelzgrpc.GetSocketRequest{SocketId: socketId}
	resp, err := clt.GetSocket(ctx, req)
	if err != nil {
		c.logger.Warn("Error getting top channels", zap.Error(err))
		return nil, err
	}
	return resp.Socket, nil
}

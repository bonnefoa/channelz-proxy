package grpc

import (
	"context"
	"strings"

	"go.uber.org/zap"
	channelzgrpc "google.golang.org/grpc/channelz/grpc_channelz_v1"
)

type ChannelResult struct {
	*channelzgrpc.Channel
	LbPolicy string
}

func extractLbPolicyFromEvents(events []*channelzgrpc.ChannelTraceEvent) string {
	for _, event := range events {
		goLbPolicyMessage := "Channel switches to new LB policy \""
		if strings.HasPrefix(event.Description, goLbPolicyMessage) {
			lbPolicy := strings.TrimRight(event.Description[len(goLbPolicyMessage):], "\"")
			return lbPolicy
		}
	}
	return ""
}

func (c *ChannelzProxyServer) GetTopChannels(ctx context.Context, address string, startChannelId int64) ([]ChannelResult, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	req := &channelzgrpc.GetTopChannelsRequest{StartChannelId: startChannelId}
	resp, err := clt.GetTopChannels(ctx, req)
	if err != nil {
		c.logger.Warn("Error getting top channels", zap.Error(err))
		return nil, err
	}
	results := make([]ChannelResult, 0)
	for _, channel := range resp.Channel {
		lbPolicy := extractLbPolicyFromEvents(channel.Data.Trace.Events)
		channelResult := ChannelResult{
			Channel:  channel,
			LbPolicy: lbPolicy,
		}
		results = append(results, channelResult)
	}
	return results, err
}

func (c *ChannelzProxyServer) GetChannel(ctx context.Context, address string, channelId int64) (*ChannelResult, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	req := &channelzgrpc.GetChannelRequest{ChannelId: channelId}
	resp, err := clt.GetChannel(ctx, req)
	if err != nil {
		c.logger.Warn("Error getting top channels", zap.Error(err))
		return nil, err
	}
	lbPolicy := extractLbPolicyFromEvents(resp.Channel.Data.Trace.Events)
	channelResult := ChannelResult{
		Channel:  resp.Channel,
		LbPolicy: lbPolicy,
	}
	return &channelResult, err
}

func (c *ChannelzProxyServer) GetSubchannel(ctx context.Context, address string, subchannelId int64) (*channelzgrpc.Subchannel, error) {
	clt, err := c.getChannelClient(address)
	if err != nil {
		return nil, err
	}
	req := &channelzgrpc.GetSubchannelRequest{SubchannelId: subchannelId}
	resp, err := clt.GetSubchannel(ctx, req)
	if err != nil {
		c.logger.Warn("Error getting top channels", zap.Error(err))
		return nil, err
	}
	return resp.Subchannel, nil
}

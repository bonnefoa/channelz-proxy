package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	channelzgrpc "google.golang.org/grpc/channelz/grpc_channelz_v1"
)

func TestExtractLbPolicy(t *testing.T) {
	events := make([]*channelzgrpc.ChannelTraceEvent, 0)
	events = append(events, &channelzgrpc.ChannelTraceEvent{
		Description: "Channel switches to new LB policy \"round_robin\"",
	})

	lbPolicy := extractLbPolicyFromEvents(events)
	assert.Equal(t, lbPolicy, "round_robin")
}

func TestGetTopChannels(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "zap")

	ctx, cancel := context.WithCancel(context.Background())
	createTestGrpcServer(t, ctx)

	c := NewChannelzProxyServer(logger)

	channels, err := c.GetTopChannels(ctx, "localhost:7654", 0)
	assert.NoError(t, err, "GetTopChannels")
	assert.Equal(t, len(channels), 1)
	cancel()
}

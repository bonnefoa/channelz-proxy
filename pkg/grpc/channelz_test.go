package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func createTestGrpcServer(t *testing.T, ctx context.Context) {
	t.Log("Start grpc server")
	channelzListener, err := net.Listen("tcp", fmt.Sprintf("localhost:7654"))
	if err != nil {
		log.Fatalf("Error listening for channelz server %v", err)
	}
	channelzServer := grpc.NewServer()
	channelz.RegisterChannelzServiceToServer(channelzServer)
	reflection.Register(channelzServer)
	go func() {
		defer channelzListener.Close()
		t.Log("Starting channelz test server")
		err = channelzServer.Serve(channelzListener)
		assert.NoError(t, err, "Channel Serve")
	}()
}

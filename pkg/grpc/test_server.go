package grpc

import (
	"context"
	"net"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/credentials/insecure"
)

func StartTestClients(ctx context.Context, logger *zap.Logger, listenAddress string) error {
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		logger.Info("Starting test client", zap.Int("num client", i))
		testConn, err := grpc.Dial(listenAddress,
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			logger.Fatal("Error opening test clients", zap.Error(err))
			return err
		}
		defer testConn.Close()
	}

	select {
	case <-ctx.Done():
		logger.Info("Context done, existing")
		return nil
	}
}

func StartTestServer(ctx context.Context, logger *zap.Logger, listenAddress string) error {
	logger = logger.Named("TestServer")
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Fatal("Failed to listen", zap.String("listenAddress", listenAddress), zap.Error(err))
		return errors.Wrap(err, "failed to listen")
	}
	defer listener.Close()
	s := grpc.NewServer()
	service.RegisterChannelzServiceToServer(s)

	logger.Info("Serving test server")

	go func() {
		err = s.Serve(listener)
		if err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()
	defer s.Stop()

	select {
	case <-ctx.Done():
		logger.Info("Context done, existing")
		return nil
	}

}

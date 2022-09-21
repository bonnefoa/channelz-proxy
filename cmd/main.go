package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bonnefoa/channelz/channelz-proxy/pkg/grpc"
	"github.com/bonnefoa/channelz/channelz-proxy/pkg/util"
	"github.com/bonnefoa/channelz/channelz-proxy/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   string
	BuildDate string
	GitCommit string
	GitBranch string
	GoVersion string

	displayVersion bool
	logLevel       *zapcore.Level

	address           string
	testServerAddress string
)

func setCliFlags() {
	logLevel = zap.LevelFlag("log-level", zap.InfoLevel, "the log level")
	flag.BoolVar(&displayVersion, "version", false, "Display version and exit")
	flag.StringVar(&address, "address", "localhost:8080", "Address for listener")
	flag.StringVar(&testServerAddress, "test-server-address", "", "Address for test grpc server")
}

func handleSignals(cancel context.CancelFunc, logger *zap.Logger) {
	sigIn := make(chan os.Signal, 100)
	signal.Notify(sigIn)
	for sig := range sigIn {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Error("Caught signal, terminating.", zap.String("Signal", sig.String()))
			cancel()
		}
	}
}

func doDisplayVersion() {
	fmt.Printf("Version: %s\n", Version)
	if GitCommit != "" {
		fmt.Printf("Git hash: %s\n", GitCommit)
	}
	if GitBranch != "" {
		fmt.Printf("Git branch: %s\n", GitBranch)
	}
	if BuildDate != "" {
		fmt.Printf("Build date: %s\n", BuildDate)
	}
	if GoVersion != "" {
		fmt.Printf("Go Version: %s\n", GoVersion)
	}
	os.Exit(0)
}

func configureLogs() *zap.Logger {
	// Configure logs
	logConfig := zap.NewProductionConfig()
	logConfig.Level.SetLevel(*logLevel)
	logger, err := logConfig.Build()
	util.FatalIf(err)

	if logConfig.Level.Level() == zap.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return logger
}

func start() {
	if displayVersion {
		doDisplayVersion()
	}
	logger := configureLogs()
	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel, logger)
	if testServerAddress != "" {
		go grpc.StartTestServer(ctx, logger, testServerAddress)
		go grpc.StartTestClients(ctx, logger, testServerAddress)
	}

	web.StartServer(ctx, address, logger)
}

func main() {
	setCliFlags()
	flag.Parse()

	start()
}

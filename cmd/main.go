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
	version   string
	buildDate string
	gitCommit string
	gitBranch string
	goVersion string

	displayVersion bool
	httpDebug      bool
	logLevel       *zapcore.Level

	listenAddress     string
	testServerAddress string
)

func setCliFlags() {
	logLevel = zap.LevelFlag("log-level", zap.InfoLevel, "the log level")
	flag.BoolVar(&displayVersion, "version", false, "Display version and exit")
	flag.BoolVar(&httpDebug, "http-debug", false, "Activate http debug")
	flag.StringVar(&listenAddress, "listen-address", "localhost:8080", "Address for listener")
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
	fmt.Printf("Version: %s\n", version)
	if gitCommit != "" {
		fmt.Printf("Git hash: %s\n", gitCommit)
	}
	if gitBranch != "" {
		fmt.Printf("Git branch: %s\n", gitBranch)
	}
	if buildDate != "" {
		fmt.Printf("Build date: %s\n", buildDate)
	}
	if goVersion != "" {
		fmt.Printf("Go Version: %s\n", goVersion)
	}
	os.Exit(0)
}

func configureLogs() *zap.Logger {
	// Configure logs
	logConfig := zap.NewProductionConfig()
	logConfig.Level.SetLevel(*logLevel)
	logger, err := logConfig.Build()
	util.FatalIf(err)

	if httpDebug {
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

	web.StartServer(ctx, listenAddress, logger)
}

func main() {
	setCliFlags()
	flag.Parse()
	start()
}

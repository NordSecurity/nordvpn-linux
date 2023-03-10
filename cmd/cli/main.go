// NordVPN command line interface application.
package main

import (
	"fmt"
	"log"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"runtime"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Salt         = ""
	Version      = ""
	Environment  = ""
	Hash         = ""
	DaemonURL    = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, internal.GetFilesharedSocket(os.Getuid()))
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(internal.UnhandledMessage)
		}
	}()

	// Setup logging
	fileLogger := &lumberjack.Logger{
		Filename:   internal.UserHomeDir() + internal.LogFilePath,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	log.SetOutput(fileLogger)

	loaderInterceptor := cli.LoaderInterceptor{}
	conn, err := grpc.Dial(
		DaemonURL,
		// Insecure credentials are OK because the connection is completely local and
		// protected by file permissions
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(loaderInterceptor.UnaryInterceptor),
		grpc.WithStreamInterceptor(loaderInterceptor.StreamInterceptor),
	)
	fileshareConn, err := grpc.Dial(
		FileshareURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(loaderInterceptor.UnaryInterceptor),
		grpc.WithStreamInterceptor(loaderInterceptor.StreamInterceptor),
	)

	cmd, err := cli.NewApp(
		Version, Environment, Hash, DaemonURL, Salt, nil, err, conn, fileshareConn, &loaderInterceptor)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	if err := cmd.Run(os.Args); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

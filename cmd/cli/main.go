// NordVPN command line interface application.
package main

import (
	"fmt"
	"log"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Salt         = ""
	Version      = "0.0.0"
	Environment  = ""
	Hash         = ""
	DaemonURL    = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, internal.GetFilesharedSocket(os.Getuid()))
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// clearFormatting removes all formatting escape sequences from input
func clearFormatting(input string) string {
	escapedString := strconv.Quote(input)
	return strings.Trim(escapedString, "\"")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(internal.UnhandledMessage)
		}
	}()

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}

	// Setup logging
	fileLogger := &lumberjack.Logger{
		Filename:   path.Join(configDir, internal.LogFilePath),
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
		Version, Environment, Hash, Salt, err, conn, fileshareConn, &loaderInterceptor)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	args := []string{}
	for _, arg := range os.Args {
		args = append(args, clearFormatting(arg))
	}

	if err := cmd.Run(args); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

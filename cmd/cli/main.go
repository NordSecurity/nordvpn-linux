// NordVPN command line interface application.
package main

import (
	"fmt"
	"log"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/clientid"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Salt        = ""
	Version     = "0.0.0"
	Environment = ""
	Hash        = ""
	DaemonURL   = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
)

func getNorduserManager() childprocess.ChildProcessManager {
	if snapconf.IsUnderSnap() {
		usr, err := user.Current()
		if err != nil {
			os.Exit(int(childprocess.CodeFailedToEnable))
		}

		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			log.Printf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}

		return process.NewNorduserGRPCProcessManager(uint32(uid))
	}

	return childprocess.NoopChildProcessManager{}
}

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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	configDir, err := internal.GetConfigDirPath(homeDir)
	if err != nil {
		log.Fatalln(err)
	}

	// Setup logging
	fileLogger := &lumberjack.Logger{
		Filename:   filepath.Join(configDir, "cli.log"),
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	log.SetOutput(fileLogger)

	clientIDMetadataInterceptor := clientid.NewInsertClientIDInterceptor(pb.ClientID_CLI)

	loaderInterceptor := cli.LoaderInterceptor{}
	conn, err := grpc.Dial(
		DaemonURL,
		// Insecure credentials are OK because the connection is completely local and
		// protected by file permissions
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(loaderInterceptor.UnaryInterceptor,
			clientIDMetadataInterceptor.SetMetadataUnaryInterceptor),
		grpc.WithChainStreamInterceptor(loaderInterceptor.StreamInterceptor,
			clientIDMetadataInterceptor.SetMetadataStreamInterceptor),
	)
	fileshareConn, err := grpc.Dial(
		fileshare_process.FileshareURL,
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

	// nolint:errcheck // we want to suppress errors in the cli app, as starting norduser is not strictly related to the
	// running command. For startup details norduser logs could be checked.
	code, err := getNorduserManager().StartProcess()
	log.Println("####################   code, err:", code, err)

	if err := cmd.Run(args); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

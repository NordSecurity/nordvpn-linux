package tray

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"

	pb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const daemonReconnectDelay = 5 * time.Second

func sniWatchState(conn *dbus.Conn, props *prop.Properties, baseIcon, connectedIcon string) {
	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	currentIcon := baseIcon

	for {
		select {
		case <-sniDone:
			return
		default:
		}

		grpcConn, err := grpc.NewClient(daemonURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Errorf("%s daemon connect failed: %v", logTag, err)
			if !sleepOrDone(daemonReconnectDelay) {
				return
			}
			continue
		}

		stream, err := pb.NewDaemonClient(grpcConn).SubscribeToStateChanges(
			context.Background(), &pb.Empty{},
		)
		if err != nil {
			_ = grpcConn.Close()
			log.Errorf("%s state subscribe failed: %v", logTag, err)
			if !sleepOrDone(daemonReconnectDelay) {
				return
			}
			continue
		}

		log.Infof("%s subscribed to daemon state changes", logTag)
		for {
			appState, err := stream.Recv()
			if err != nil {
				log.Errorf("%s state stream error: %v", logTag, err)
				break
			}
			cs, ok := appState.GetState().(*pb.AppState_ConnectionStatus)
			if !ok {
				continue
			}
			icon := baseIcon
			if cs.ConnectionStatus.GetState() == pb.ConnectionState_CONNECTED {
				icon = connectedIcon
			}
			if icon != currentIcon {
				currentIcon = icon
				sniUpdateIcon(conn, props, icon)
			}
		}

		_ = grpcConn.Close()
		if !sleepOrDone(daemonReconnectDelay) {
			return
		}
	}
}

func sniUpdateIcon(conn *dbus.Conn, props *prop.Properties, iconName string) {
	props.SetMust(sniIface, "IconName", iconName)
	if err := conn.Emit(sniPath, sniIface+".NewIcon"); err != nil {
		log.Errorf("%s emit NewIcon: %v", logTag, err)
	}
	log.Infof("%s icon updated to %s", logTag, iconName)
}

// sleepOrDone waits for d or until sniDone is closed. Returns false if sniDone fired.
func sleepOrDone(d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-sniDone:
		return false
	}
}

package clientid

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	clientIDMetadataKey = "client-id"

	connectItemName    = "connect"
	disconnectItemName = "disconnect"
	loginItemName      = "login"
	logoutItemName     = "logout"

	cliClientString  = "cli"
	guiClientString  = "gui"
	trayClientString = "tray"
)

type CliendIDMiddleware struct {
	publisher events.Publisher[events.UiItemsAction]
}

func NewClientIDMiddleware(publisher events.Publisher[events.UiItemsAction]) CliendIDMiddleware {
	return CliendIDMiddleware{
		publisher: publisher,
	}
}

func clientIDToClientString(clientID pb.ClientID) string {
	switch clientID {
	case pb.ClientID_CLI:
		return cliClientString
	case pb.ClientID_GUI:
		return guiClientString
	case pb.ClientID_TRAY:
		return trayClientString
	case pb.ClientID_UNKNOWN_CLIENT:
		return ""
	}
	return ""
}

func fullMethodNameToItemName(fullMethodName string) string {
	switch fullMethodName {
	case pb.Daemon_Connect_FullMethodName:
		return connectItemName
	case pb.Daemon_Disconnect_FullMethodName:
		return disconnectItemName
	case pb.Daemon_LoginOAuth2_FullMethodName:
		return loginItemName
	case pb.Daemon_Logout_FullMethodName:
		return logoutItemName
	default:
		return ""
	}
}

func parseClientID(clientID string) (string, error) {
	clientIDInt, err := strconv.ParseInt(clientID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("parsing client ID: %w", err)
	}

	clientString := clientIDToClientString(pb.ClientID(clientIDInt))
	if clientString == "" {
		return "", fmt.Errorf("unknown client ID: %s", clientID)
	}

	return clientString, nil
}

func (c *CliendIDMiddleware) notifyAboutClickEvent(ctx context.Context, fullMethod string) {
	itemName := fullMethodNameToItemName(fullMethod)
	if itemName == "" {
		return
	}

	clientString := ""
	var md metadata.MD
	var ok bool
	if md, ok = metadata.FromIncomingContext(ctx); !ok {
		return
	}

	var metadata []string
	if metadata = md.Get(clientIDMetadataKey); len(metadata) == 0 {
		return
	}

	clientString, err := parseClientID(metadata[0])
	if err != nil {
		log.Println(internal.ErrorPrefix, "parsing client ID:", err)
		return
	}

	c.publisher.Publish(events.UiItemsAction{
		ItemName:      itemName,
		ItemType:      "button",
		FormReference: clientString,
	})
}

func (c *CliendIDMiddleware) UnaryMiddleware(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo) (interface{}, error) {
	c.notifyAboutClickEvent(ctx, info.FullMethod)

	return nil, nil
}

func (c *CliendIDMiddleware) StreamMiddleware(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo) error {
	c.notifyAboutClickEvent(ss.Context(), info.FullMethod)

	return nil
}

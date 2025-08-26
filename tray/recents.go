package tray

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"google.golang.org/grpc"
)

const (
	// Timeouts
	recentConnectionsTimeout = 2 * time.Second
)

func makeDisplayLabel(conn *pb.RecentConnectionModel) string {
	switch conn.ConnectionType {
	case config.ServerSelectionRule_CITY:
		if conn.Country != "" && conn.City != "" {
			return fmt.Sprintf("%s, %s", conn.Country, conn.City)
		}
		return ""

	case config.ServerSelectionRule_COUNTRY:
		return conn.Country

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		return conn.SpecificServerName

	case config.ServerSelectionRule_GROUP:
		return strings.ReplaceAll(conn.Group.String(), "_", " ")

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		if conn.Group != config.ServerGroup_UNDEFINED && conn.Country != "" {
			group := strings.ReplaceAll(conn.Group.String(), "_", " ")
			return fmt.Sprintf("%s (%s)", group, conn.Country)
		}
		return ""

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		if conn.Group != config.ServerGroup_UNDEFINED {
			group := strings.ReplaceAll(conn.Group.String(), "_", " ")
			if conn.Country != "" && conn.City != "" {
				return fmt.Sprintf("%s (%s, %s)", group, conn.Country, conn.City)
			} else if conn.Country != "" {
				return fmt.Sprintf("%s (%s)", group, conn.Country)
			}
		}
		return ""

	case config.ServerSelectionRule_NONE, config.ServerSelectionRule_RECOMMENDED:
		return ""
	}

	return ""
}

func fetchRecentConnections(ti *Instance) []*pb.RecentConnectionModel {
	ctx, cancel := context.WithTimeout(context.Background(), recentConnectionsTimeout)
	defer cancel()

	resp, err := ti.client.GetRecentConnections(
		ctx,
		&pb.RecentConnectionsRequest{Limit: maxRecentConnections},
		grpc.WaitForReady(true),
	)

	if err != nil || resp == nil {
		return nil
	}

	return resp.Connections
}

func connectByConnectionModel(ti *Instance, model *pb.RecentConnectionModel) bool {
	if model == nil {
		return false
	}

	switch model.ConnectionType {
	case config.ServerSelectionRule_RECOMMENDED:
		return ti.connect("", "")

	case config.ServerSelectionRule_CITY:
		city_str := strings.ReplaceAll(model.City, " ", "_")
		return ti.connect(city_str, "")

	case config.ServerSelectionRule_COUNTRY:
		country_str := strings.ReplaceAll(model.Country, " ", "_")
		return ti.connect(country_str, "")

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		return ti.connect(model.SpecificServer, "")

	case config.ServerSelectionRule_GROUP:
		group_str := strings.ReplaceAll(model.Group.String(), "_", " ")
		return ti.connect("", group_str)

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		group_str := strings.ReplaceAll(model.Group.String(), "_", " ")
		country_str := strings.ReplaceAll(model.Country, " ", "_")
		return ti.connect(country_str, group_str)

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		group_str := strings.ReplaceAll(model.Group.String(), "_", " ")
		return ti.connect(model.SpecificServer, group_str)

	case config.ServerSelectionRule_NONE:
		return false
	}

	return false
}

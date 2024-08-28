#!/bin/bash
set -euxo pipefail

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/protocol.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/technology.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/group.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/account.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/cities.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/common.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/connect.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/countries.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/logout.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login_with_token.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/ping.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/plans.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/rate.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/register.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/set.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/settings.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/status.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/token.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/purchase.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/empty.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/fsnotify.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/service_response.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/peer.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/invite.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/transfer.proto -I protobuf/fileshare
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/fileshare.proto -I protobuf/fileshare
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/snapconf/snapconf.proto -I protobuf/snapconf
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/norduser/norduser.proto -I protobuf/norduser
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/state.proto -I protobuf/daemon

protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/daemon/service.proto -I protobuf/daemon
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/meshnet/service.proto -I protobuf/meshnet
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/fileshare/service.proto -I protobuf/fileshare
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/norduser/service.proto -I protobuf/norduser

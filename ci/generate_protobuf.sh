#!/bin/bash
set -euxo pipefail

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/protocol.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/technology.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/group.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/account.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/cities.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/common.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/connect.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/countries.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/logout.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login_with_token.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/ping.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/rate.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/set.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/settings.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/status.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/token.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/purchase.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/servers.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/state.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/nordwhisper_enabled.proto

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/empty.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/fsnotify.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/service_response.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/peer.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/invite.proto

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/transfer.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/fileshare.proto

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/snapconf/snapconf.proto

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/norduser/norduser.proto

protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/daemon/service.proto
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/meshnet/service.proto
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/fileshare/service.proto
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/norduser/service.proto

outDir="${PWD}"/test/qa/lib
mkdir -p "${outDir}"
touch "${outDir}"/protobuf/daemon/__init__.py

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}" \
	--python_out="${outDir}" \
	--pyi_out="${outDir}" \
	"${PWD}"/protobuf/daemon/*.proto \
	"${PWD}"/protobuf/daemon/config/*.proto

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}" \
	--grpc_python_out="${outDir}" \
	"${PWD}"/protobuf/daemon/service.proto

touch "${outDir}"/protobuf/meshnet/__init__.py

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}" \
	--python_out="${outDir}" \
	--pyi_out="${outDir}" \
	"${PWD}"/protobuf/meshnet/*.proto

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}" \
	--grpc_python_out="${outDir}" \
	"${PWD}"/protobuf/meshnet/service.proto

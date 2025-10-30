#!/bin/bash
set -euxo pipefail

# ================================  [ Core ] ================================

protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/protocol.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/technology.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/analytics_consent.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/config/group.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/account.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/cities.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/common.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/connect.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/countries.proto
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/logout.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/login_with_token.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/ping.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/rate.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/set.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/settings.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/status.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/token.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/purchase.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/servers.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/state.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/features.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/empty.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/fsnotify.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/service_response.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/peer.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/meshnet/invite.proto -I protobuf/meshnet
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/transfer.proto -I protobuf/fileshare
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/fileshare/fileshare.proto -I protobuf/fileshare
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/snapconf/snapconf.proto -I protobuf/snapconf
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/norduser/norduser.proto -I protobuf/norduser
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/defaults.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/telemetry/v1/fields.proto -I protobuf/daemon/telemetry/v1
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/recent_connections.proto -I protobuf/daemon
protoc --go_opt=module=github.com/NordSecurity/nordvpn-linux --go_out=. protobuf/daemon/server_selection_rule.proto -I protobuf/daemon

protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/daemon/service.proto -I protobuf/daemon
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/meshnet/service.proto -I protobuf/meshnet
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/fileshare/service.proto -I protobuf/fileshare
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/norduser/service.proto -I protobuf/norduser
protoc --go_grpc_opt=module=github.com/NordSecurity/nordvpn-linux --go_grpc_out=. protobuf/daemon/telemetry/v1/service.proto -I protobuf/daemon/telemetry/v1

# ================================  [ Python Tests ] ================================

outDir="${PWD}"/test/qa/lib/protobuf/daemon
mkdir -p "${outDir}"
touch "${outDir}"/__init__.py

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}"/protobuf/daemon \
	--python_out="${outDir}" \
	--pyi_out="${outDir}" \
	"${PWD}"/protobuf/daemon/*.proto \
	"${PWD}"/protobuf/daemon/config/*.proto

python3 -m grpc_tools.protoc \
	--proto_path="${PWD}"/protobuf/daemon \
	--grpc_python_out="${outDir}" \
	"${PWD}"/protobuf/daemon/service.proto

# ================================  [ GUI ] ================================

OUT="${PWD}/gui/lib/pb"
rm -fr "${OUT}"
mkdir -p "${OUT}"

# There is a problem with dart and google protobuf https://github.com/google/protobuf.dart/issues/483
# Generate once google protobufs and then create symbolic links into each folder.
# Alternative is to let protoc generate the folders multiple times automatically
echo "Generate google protobufs using $(dart --version)"
GOOGLE_PROTO_ROOT_DIR="/usr/include"
protoc -I="${GOOGLE_PROTO_ROOT_DIR}" --dart_out="${OUT}" "${GOOGLE_PROTO_ROOT_DIR}/google/protobuf/timestamp.proto"

# used GRPC files
GRPC_FILES=(
    "protobuf/daemon/service.proto"
    "protobuf/meshnet/service.proto"
    "protobuf/fileshare/service.proto"
)
echo "**** Generate GRPC files *****"
for i in "${GRPC_FILES[@]}"; do
    DIR=$(dirname "$i")
    FILE=$(basename "$i")
    pushd "${DIR}" > /dev/null
    DIR_NAME_WITHOUT_PROTOBUF=$(echo "${DIR}" | cut -d'/' -f2-)
    PROTO_OUT="${OUT}/${DIR_NAME_WITHOUT_PROTOBUF}"
    echo "* $i -> ${PROTO_OUT}"

    mkdir -p "${PROTO_OUT}"
    protoc --dart_out=grpc:"${PROTO_OUT}" "${FILE}"
    popd >/dev/null
done

PROTO_FILES=$(find ./protobuf -name "*.proto" ! -name service.proto -not -path "*/snapconf/*" -not -path "*/norduser/*" -not -path "*/libtelio/*" -not -path "*/parts/*")

echo "**** Generate files *****"
for i in ${PROTO_FILES}; do
    DIR=$(dirname "$i")
    FILE=$(basename "$i")
    DIR_NAME_WITHOUT_PROTOBUF=$(echo "${DIR}" | cut -d'/' -f3-)
    PROTO_OUT="${OUT}/${DIR_NAME_WITHOUT_PROTOBUF}"

    FILENAME="${FILE%.*}"
    echo "* ${FILE} -> ${PROTO_OUT}/${FILENAME}.pb.dart"
    if [ -f "${PROTO_OUT}/${FILENAME}.pb.dart" ]; then
        echo "---> Skipping $i, already created"
        continue
    fi

    pushd "${DIR}" > /dev/null
    mkdir -p "${PROTO_OUT}"

    # create a symlink for google files, not to create it each time, for each folder.
    # this can be removed when https://github.com/google/protobuf.dart/issues/483 is fixed
    if [ ! -L "${PROTO_OUT}/google" ]; then
        ln -s "../google" "${PROTO_OUT}/google"
    fi

    # using protoc --dart_out="$PROTO_OUT" "$FILE" "$GOOGLE_PROTO_ROOT_DIR/google/protobuf/timestamp.proto" creates google multiple times
    protoc --dart_out="${PROTO_OUT}" "${FILE}"

    find "$OUT" -type f -iname "*pbserver.dart" -delete

    popd > /dev/null
done

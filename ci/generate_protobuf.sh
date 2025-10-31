#!/bin/bash
set -euxo pipefail

# ================================[ Declaration ]================================

function for_core() {
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

}

function for_python_tests() {
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
}

function for_gui() {
  local out="${PWD}/gui/lib/pb"
  rm -fr "${out}"
  mkdir -p "${out}"

  # There is a problem with dart and google protobuf https://github.com/google/protobuf.dart/issues/483
  # Generate once google protobufs and then create symbolic links into each folder.
  # Alternative is to let protoc generate the folders multiple times automatically
  echo "Generate google protobufs using $(dart --version)"
  local google_proto_root_dir="/usr/include"
  protoc -I="${google_proto_root_dir}" --dart_out="${out}" "${google_proto_root_dir}/google/protobuf/timestamp.proto"

  # used GRPC files
  local grpc_files=(
    "protobuf/daemon/service.proto"
    "protobuf/meshnet/service.proto"
    "protobuf/fileshare/service.proto"
  )
  echo "**** Generate GRPC files *****"

  local dir
  local file
  local dir_name_without_protobuf
  local proto_out
  for i in "${grpc_files[@]}"; do
    dir=$(dirname "${i}")
    file=$(basename "${i}")
    pushd "${dir}" >/dev/null
    dir_name_without_protobuf=$(echo "${dir}" | cut -d'/' -f2-)
    proto_out="${out}/${dir_name_without_protobuf}"
    echo "* ${i} -> ${proto_out}"

    mkdir -p "${proto_out}"
    protoc --dart_out=grpc:"${proto_out}" "${file}"
    popd >/dev/null
  done

  local proto_files
  proto_files=$(find ./protobuf -name "*.proto" ! -name service.proto -not -path "*/snapconf/*" -not -path "*/norduser/*" -not -path "*/libtelio/*" -not -path "*/parts/*")

  echo "**** Generate files *****"
  local filename
  for i in ${proto_files}; do
    dir=$(dirname "${i}")
    file=$(basename "${i}")
    dir_name_without_protobuf=$(echo "${dir}" | cut -d'/' -f3-)
    proto_out="${out}/${dir_name_without_protobuf}"

    filename="${file%.*}"
    echo "* ${file} -> ${proto_out}/${filename}.pb.dart"
    if [ -f "${proto_out}/${filename}.pb.dart" ]; then
      echo "---> Skipping ${i}, already created"
      continue
    fi

    pushd "${dir}" >/dev/null
    mkdir -p "${proto_out}"

    # create a symlink for google files, not to create it each time, for each folder.
    # this can be removed when https://github.com/google/protobuf.dart/issues/483 is fixed
    if [ ! -L "${proto_out}/google" ]; then
      ln -s "../google" "${proto_out}/google"
    fi

    # using protoc --dart_out="${proto_out}" "${file}" "${google_proto_root_dir}/google/protobuf/timestamp.proto" creates google multiple times
    protoc --dart_out="${proto_out}" "${file}"

    find "${out}" -type f -iname "*pbserver.dart" -delete

    popd >/dev/null
  done
}

function generate() {
  for_core
  for_python_tests
  for_gui
}

# ================================[ Execution ]================================

generate

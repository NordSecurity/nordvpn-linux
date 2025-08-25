#!/usr/bin/env bash
set -e 
OUT="${PWD}/lib/pb"
echo $OUT
rm -fr "$OUT"
mkdir -p "$OUT"

# There is a problem with dart and google protobuf https://github.com/google/protobuf.dart/issues/483
# Generate once google protobufs and then create symbolic links into each folder.
# Alternative is to let protoc generate the folders multiple times automatically
echo Generate google protobufs using `dart --version`
GOOGLE_PROTO_ROOT_DIR="/usr/include"
for FILE in "timestamp.proto"; do
    protoc -I="$GOOGLE_PROTO_ROOT_DIR" --dart_out="$OUT" "$GOOGLE_PROTO_ROOT_DIR/google/protobuf/$FILE"
done

pushd nordvpn-linux > /dev/null
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
    pushd $DIR > /dev/null
    DIR_NAME_WITHOUT_PROTOBUF=$(echo $DIR | cut -d'/' -f2-)
    PROTO_OUT="$OUT/$DIR_NAME_WITHOUT_PROTOBUF"
    echo "* $i -> $PROTO_OUT"

    mkdir -p "$PROTO_OUT"
    protoc --dart_out=grpc:"$PROTO_OUT" "$FILE"
    popd >/dev/null
done

echo "**** Generate files *****"
for i in $(find . -name "*.proto" ! -name service.proto  -not -path "*/snapconf/*" -not -path "*/norduser/*" ); do
    DIR=$(dirname "$i")
    FILE=$(basename "$i")
    DIR_NAME_WITHOUT_PROTOBUF=$(echo $DIR | cut -d'/' -f3-)
    PROTO_OUT="$OUT/$DIR_NAME_WITHOUT_PROTOBUF"

    FILENAME="${FILE%.*}"
    echo "* $FILE -> $PROTO_OUT/$FILENAME.pb.dart" 
    if [ -f "$PROTO_OUT/$FILENAME.pb.dart" ]; then
        echo "---> Skipping $i, already created"
        continue
    fi

    pushd $DIR > /dev/null
    mkdir -p "$PROTO_OUT"

    # create a symlink for google files, not to create it each time, for each folder.
    # this can be removed when https://github.com/google/protobuf.dart/issues/483 is fixed
    if [ ! -L "$PROTO_OUT/google" ]; then
        ln -s "../google" "$PROTO_OUT/google"
    fi
    
    # using protoc --dart_out="$PROTO_OUT" "$FILE" "$GOOGLE_PROTO_ROOT_DIR/google/protobuf/timestamp.proto" creates google multiple times
    protoc --dart_out="$PROTO_OUT" "$FILE"

    find "$OUT" -type f -iname "*pbserver.dart" -delete

    popd > /dev/null
done

popd > /dev/null
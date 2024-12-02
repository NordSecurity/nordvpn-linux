FROM ghcr.io/nordsecurity/nordvpn-linux/golang:ubuntu-20.04-20240725 AS builder

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN wget --progress=dot:giga https://github.com/protocolbuffers/protobuf/releases/download/v21.6/protoc-21.6-linux-x86_64.zip -O /tmp/protoc.zip && \
    echo "6a9fc36363a2d05d73fc363a46cd57d849068d33305db39f77daac8ba073e818 /tmp/protoc.zip" | sha256sum -c - && \
    unzip /tmp/protoc.zip -d /tmp/protoc && rm /tmp/protoc.zip && \
    # Install protoc-gen-go and protoc-gen-go-grpc
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 && \
    ln -s /usr/bin/protoc-gen-go-grpc /usr/bin/protoc-gen-go_grpc && \
    go clean -cache -modcache

FROM alpine:3.20.3
COPY --from=builder /usr/bin/protoc-gen-go /usr/bin/
COPY --from=builder /usr/bin/protoc-gen-go_grpc /usr/bin/
COPY --from=builder /tmp/protoc /usr/
COPY requirements.txt /tmp/requirements.txt

RUN apk add --no-cache bash libc6-compat libstdc++ python3 py3-pip
RUN python3 -m pip install --break-system-packages --no-cache-dir -r /tmp/requirements.txt && \
	rm -rf /tmp/*

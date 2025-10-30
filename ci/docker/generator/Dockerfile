FROM ghcr.io/nordsecurity/nordvpn-linux/golang:debian-bullseye-20250627 AS builder

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN wget --progress=dot:giga https://github.com/protocolbuffers/protobuf/releases/download/v21.6/protoc-21.6-linux-x86_64.zip -O /tmp/protoc.zip && \
    echo "6a9fc36363a2d05d73fc363a46cd57d849068d33305db39f77daac8ba073e818 /tmp/protoc.zip" | sha256sum -c - && \
    unzip /tmp/protoc.zip -d /tmp/protoc && rm /tmp/protoc.zip && \
    # Install protoc-gen-go and protoc-gen-go-grpc
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 && \
    ln -s /usr/bin/protoc-gen-go-grpc /usr/bin/protoc-gen-go_grpc && \
    go clean -cache -modcache

FROM dart:3.8.1 AS dart-builder

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates unzip python3 python3-pip bash libc6 libstdc++6 \
  && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/bin/protoc-gen-go /usr/bin/
COPY --from=builder /usr/bin/protoc-gen-go_grpc /usr/bin/
COPY --from=builder /tmp/protoc /usr/

COPY --from=dart-builder /usr/lib/dart /usr/lib/dart

COPY requirements.txt /tmp/requirements.txt

RUN python3 -m pip install --break-system-packages --no-cache-dir -r /tmp/requirements.txt && \
	rm -rf /tmp/*

ENV PATH="${PATH}:/usr/lib/dart/bin"

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} generator && useradd -l -m -u ${USER_ID} -g generator generator

USER generator

ENV PATH="${PATH}:/home/generator/.pub-cache/bin"

RUN dart pub global activate protoc_plugin 22.5.0


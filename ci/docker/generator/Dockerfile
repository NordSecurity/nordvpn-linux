FROM golang:1.20.1-buster

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ENV GOBIN=/usr/bin

RUN apt-get update \
    && apt-get install -y zip \
    # cleanup
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*


RUN curl -L https://github.com/protocolbuffers/protobuf/releases/download/v21.6/protoc-21.6-linux-x86_64.zip -o /tmp/protoc.zip \
    && unzip /tmp/protoc.zip -d /usr/

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 \
    && ln -s /usr/bin/protoc-gen-go-grpc /usr/bin/protoc-gen-go_grpc

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN groupadd -g ${GROUP_ID} generator && useradd -l -m -u ${USER_ID} -g generator generator && echo "generator ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER generator

CMD ["exec", "$@"]

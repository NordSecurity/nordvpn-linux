FROM golangci/golangci-lint:v1.51.2

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN apt update && apt install -y libxml2-dev \
    && groupadd -g ${GROUP_ID} linter && useradd -l -m -u ${USER_ID} -g linter linter

USER linter

CMD ["exec", "$@"]

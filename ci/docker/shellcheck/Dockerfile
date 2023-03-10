FROM alpine:latest

LABEL org.opencontainers.image.source=https://github.com/NordSecurity/nordvpn-linux

RUN tag="v0.9.0"; \
  url_base='https://github.com/koalaman/shellcheck/releases/download'; \
  tar_file="${tag}/shellcheck-${tag}.linux.x86_64.tar.xz"; \
  wget "${url_base}/${tar_file}" -O - | tar xJf -; \
  mv "shellcheck-${tag}/shellcheck" /bin/; \
  rm -rf "shellcheck-${tag}"; \
  ls -laF /bin/shellcheck

CMD ["exec", "$@"]

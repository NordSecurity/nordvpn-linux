# In order to build this image run

```bash
docker build -t ghcr.io/nordsecurity/nordvpn-linux/ruster:<version> \
  --build-arg GL_ACCESS_TOKEN=<gitlab-access-token> \
  --build-arg SQLITE_DOWNLOAD_URL_PREFIX=<url-prefix> .
```

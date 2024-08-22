# In order to build this image run

```bash
docker build -t ghcr.io/nordsecurity/nordvpn-linux/ruster:<version> \
  --build-arg SQLITE_DOWNLOAD_URL_PREFIX=<url-prefix> \
  --secret id=gl_access_token,src=<path-to-file-with-gitlab-access-token> .
```

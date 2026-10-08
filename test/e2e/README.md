# Some info

## Building the image

So this thing - we probably won't need it, we can build and manage images
on our own. But leaving it here for a reference.

```sh
go run ./cmd/build-fedora-image -out ~/vms/fedora-44-nordvpn.qcow2
```

## Running

```sh
NORDVPN_E2E_IMAGE=<path-to-build-qcow2-image> NORDVPN_TOKEN=... \
  go test -v -count=1 -timeout 30m ./vmtests/
```

module github.com/NordSecurity/nordvpn-linux

go 1.25.3

// Bindings
// NOTE: If you are chaning the binding versions here, keep in mind that you
// may also need to update versions in `./lib-versions.env` file.
require (
	github.com/NordSecurity/libdrop-go/v9 v9.0.0
	github.com/NordSecurity/libtelio-go/v6 v6.2.3
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/NordSecurity/gopenvpn v0.0.0-20230117114932-2252c52984b4
	github.com/NordSecurity/systray v0.0.0-20260618073639-14a79f2708b4
	github.com/coreos/go-semver v0.3.1
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/docker/docker v28.5.1+incompatible
	github.com/docker/go-units v0.5.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/esiqveland/notify v0.13.2
	github.com/fatih/color v1.15.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-co-op/gocron/v2 v2.5.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/google/go-cmp v0.7.0
	github.com/google/nftables v0.3.0
	github.com/google/uuid v1.6.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/magefile/mage v1.14.0
	github.com/miekg/dns v1.1.72
	github.com/milosgajdos/tenus v0.0.3
	github.com/pmezard/go-difflib v1.0.0
	github.com/quic-go/quic-go v0.57.0
	github.com/snapcore/snapd v0.0.0-20260619062016-77ff61930ffd
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v2 v2.25.0
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.46.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	golang.org/x/mod v0.31.0
	golang.org/x/net v0.48.0
	golang.org/x/sys v0.39.0
	golang.org/x/term v0.38.0
	golang.org/x/text v0.32.0
	golang.zx2c4.com/wireguard v0.0.0-20230313165553-0ad14a89f5f9
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.10
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gotest.tools/v3 v3.4.0
)

require (
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/canonical/cpuid v0.0.0-20220614022739-219e067757cb // indirect
	github.com/canonical/go-efilib v1.8.0 // indirect
	github.com/canonical/go-kbkdf v0.0.0-20250104172618-3b1308f9acf9 // indirect
	github.com/canonical/go-password-validator v0.0.0-20250617132709-1b205303ca54 // indirect
	github.com/canonical/go-sp800.90a-drbg v0.0.0-20210314144037-6eeb1040d6c3 // indirect
	github.com/canonical/go-tpm2 v1.16.2 // indirect
	github.com/canonical/tcglog-parser v0.0.0-20240924110432-d15eaf652981 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.3 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/libcontainer v2.2.1+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mdlayher/netlink v1.7.3-0.20250113171957-fbb4dce95f42 // indirect
	github.com/mdlayher/socket v0.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pilebones/go-udev v0.9.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/snapcore/secboot v0.0.0-20260424115705-c00dcfff2f83 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	maze.io/x/crypto v0.0.0-20190131090603-9b94c9afe066 // indirect
)

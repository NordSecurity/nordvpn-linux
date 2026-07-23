module github.com/NordSecurity/nordvpn-linux

go 1.26.3

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
	github.com/deckarep/golang-set/v2 v2.9.0
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-units v0.5.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/esiqveland/notify v0.14.0
	github.com/fatih/color v1.19.0
	github.com/fsnotify/fsnotify v1.10.1
	github.com/go-co-op/gocron/v2 v2.22.0
	github.com/godbus/dbus/v5 v5.2.2
	github.com/google/go-cmp v0.7.0
	github.com/google/nftables v0.3.0
	github.com/google/uuid v1.6.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/magefile/mage v1.17.2
	github.com/miekg/dns v1.1.72
	github.com/milosgajdos/tenus v0.0.3
	github.com/pmezard/go-difflib v1.0.0
	github.com/quic-go/quic-go v0.60.0
	github.com/snapcore/snapd v0.0.0-20260720113602-8bbab2efb256
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v2 v2.27.7
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.54.0
	golang.org/x/exp v0.0.0-20260718201538-764159d718ef
	golang.org/x/mod v0.38.0
	golang.org/x/net v0.57.0
	golang.org/x/sys v0.47.0
	golang.org/x/term v0.45.0
	golang.org/x/text v0.40.0
	golang.zx2c4.com/wireguard v0.0.0-20260522210424-ecfc5a8d5446
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gotest.tools/v3 v3.4.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
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
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.7.0 // indirect
	github.com/docker/libcontainer v2.2.1+incompatible // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	github.com/mdlayher/netlink v1.11.2 // indirect
	github.com/mdlayher/socket v0.6.1 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pilebones/go-udev v0.9.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/snapcore/secboot v0.0.0-20260623135244-457b03a16d19 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260715232425-e75dac1f907d // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	maze.io/x/crypto v0.0.0-20190131090603-9b94c9afe066 // indirect
)

module github.com/NordSecurity/nordvpn-linux

go 1.22.1

toolchain go1.22.2

// Bindings
// NOTE: If you are chaning the binding versions here, keep in mind that you
// may also need to update versions in `./lib-versions.env` file.
require (
	github.com/NordSecurity/libdrop-go/v8 v8.0.0-20241017064027-670787595588
	github.com/NordSecurity/libtelio-go/v5 v5.1.5
)

require (
	github.com/NordSecurity/gopenvpn v0.0.0-20230117114932-2252c52984b4
	github.com/NordSecurity/systray v0.0.0-20240327004800-3e3b59c1b83d
	github.com/coreos/go-semver v0.3.1
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/docker/docker v25.0.6+incompatible
	github.com/docker/go-units v0.5.0
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/esiqveland/notify v0.13.2
	github.com/fatih/color v1.15.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-co-op/gocron/v2 v2.5.0
	github.com/go-ping/ping v1.1.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/jbowtie/gokogiri v0.0.0-20190301021639-37f655d3078f
	github.com/jbowtie/ratago v0.0.0-20200401224626-3140c0a9b186
	github.com/magefile/mage v1.14.0
	github.com/milosgajdos/tenus v0.0.3
	github.com/quic-go/quic-go v0.48.2
	github.com/stretchr/testify v1.9.0
	github.com/urfave/cli/v2 v2.25.0
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/crypto v0.31.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	golang.org/x/mod v0.17.0
	golang.org/x/net v0.33.0
	golang.org/x/oauth2 v0.18.0
	golang.org/x/sys v0.28.0
	golang.org/x/term v0.27.0
	golang.org/x/text v0.21.0
	golang.zx2c4.com/wireguard v0.0.0-20230313165553-0ad14a89f5f9
	google.golang.org/api v0.169.0
	google.golang.org/grpc v1.64.1
	google.golang.org/protobuf v1.34.2
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gotest.tools/v3 v3.4.0
)

require (
	cloud.google.com/go/compute v1.25.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/libcontainer v2.2.1+incompatible // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

# daemon/

The privileged, long-running service. It owns all VPN state; the CLI, GUI and tray are
unprivileged clients that reach it over gRPC on a unix socket. Anything here runs as root.

## Generated code

- `daemon/pb/**` is generated from `protobuf/daemon/*.proto`. Never edit it.
- Regenerate with `mage generate:protobufDocker` (runs `ci/generate_protobuf.sh`).
- The `.proto` files are a public API contract with clients that may be older than the daemon.
  Adding fields and RPCs is safe; renaming, renumbering or removing them is a compatibility
  decision, not a refactor.

## RPC handlers

- One RPC per file, named `rpc_<snake_case_name>.go`, as a method on `*RPC`.
- Register new services in `cmd/daemon/main.go` (`pb.RegisterDaemonServer`).
- Most handlers report failures **in band** — a payload carrying `internal.Code*`
  (`internal/codes.go`) with a `nil` error — because clients switch on that code. Some handlers
  return a gRPC error instead. Follow whatever the handler you are editing already does; do not
  change an existing handler's error style, it is client-visible behaviour.
- Read configuration through the injected `config.Manager` (`r.cm.Load(&cfg)`), never by touching
  config files directly.

## Logging

- The stdlib `log` package is banned outside `log/` and depguard will fail the lint run. Use
  `github.com/NordSecurity/nordvpn-linux/log`.
- Prefer the subsystem loggers in `log/prefixes.go` (`log.DNS`, `log.Recents`, `log.Telemetry`, …)
  over bare `log.Error` when the code belongs to one of those areas.
- Never log tokens, credentials, session data or server credentials. Logs ship in support bundles.

## Running as root

- Use the `internal.Perm*` constants (`internal/constants.go`) for any file or socket you create;
  do not write octal literals.
- Treat every value that originated in a client request as untrusted: validate paths, never
  interpolate them into shell invocations, and prefer `exec.Command` with an explicit argv.
- `mage test:gosec` runs in CI over this tree; a new finding blocks the build.

## Analytics and consent

- Telemetry and moose events must stay behind the consent state
  (`daemon/userconsent.go`, `daemon/telemetry/`). Do not add an event path that reports before
  `AnalyticsConsentChecker` says the flow is complete.

## Build tags

- `moose`, `telio`, `quench`, `drop` gate optional native dependencies. Code must compile both
  with and without them — guard tag-specific code in separate files, not with `if` statements.
- Unit tests build with `-tags internal`.

## Tests

- Every test starts with `category.Set(t, category.<Category>)` — see `test/category/category.go`.
- The default unit run excludes `root,link,firewall,route,file,integration,notworking`, so a test
  that needs privileges, real network links or a live backend must be categorised accordingly or
  it will run in the wrong environment.
- Golden files go through `test/golden` (`golden.AssertMatchesGolden`).

## Commands

| Task | Command |
|---|---|
| unit tests | `mage test:go` (wraps `ci/test.sh`) |
| full suite incl. cgo | `mage test:cgoDocker` |
| lint | `mage test:lint` (config: `.golangci-lint.yml`) |
| security scan | `mage test:gosec` |
| regenerate protobuf | `mage generate:protobufDocker` |

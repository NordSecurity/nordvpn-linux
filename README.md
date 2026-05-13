# NordVPN for Linux

![icon](./assets/icon.svg)

---

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/NordSecurity/nordvpn-linux/badge)](https://scorecard.dev/viewer/?uri=github.com/NordSecurity/nordvpn-linux)

The [NordVPN](https://nordvpn.com/features/) Linux application provides a
simple and user-friendly command line interface for accessing all the
different features of NordVPN. Users can choose from a list of server
locations around the world, or let the application automatically select
the best server for them. They can also customize their connection
settings, such as choosing a specific protocol or enabling the kill
switch feature.

The application manages:

- network interfaces using
  [WireGuard](https://www.wireguard.com/) (NordLynx) and tun (OpenVPN),
- firewall with the help of
  [iptables](https://www.netfilter.org/projects/iptables/index.html),
- routing via the
  [netlink](https://www.man7.org/linux/man-pages/man7/netlink.7.html)
  kernel interface and
- DNS using systemd-resolved, resolvconf, or NetworkManager depending
  on what is available on the system.

---

## Versioning

The project follows [semver](https://semver.org/). Version tags and
release branches must be named accordingly.

## Contributing

We are happy to accept contributions for the project. Please check out
[CONTRIBUTE.md](./CONTRIBUTE.md) for more details on how to do so.

## Building

You can find everything related to building, testing and environment
setup in [BUILD.md](./BUILD.md).

## Troubleshooting

### Log level

The log verbosity of the NordVPN daemons can be changed at runtime
without restarting by writing to:

- `/run/nordvpn/loglevel` for DEB and RPM
- `/var/snap/nordvpn/common/run/nordvpn/loglevel` for SNAP

Example:

```sh
echo "debug" | sudo tee /run/nordvpn/loglevel
```

Valid values are `debug`, `info`, `warn`, `error`, `fatal` and `off`.

## Installing

For installing an already released version please follow the
instructions on our
[official page](https://nordvpn.com/download/linux/#install-nordvpn).

### Supported distros

<https://nordvpn.com/download/linux/>

Distributions are not supported after their end of life.

This project is licensed under the terms of the
[GNU General Public License v3.0](./LICENSE.md) only.
The registered trademark Linux® is used pursuant to a sublicense from
the Linux Foundation, the exclusive licensee of Linus Torvalds, owner
of the mark on a world-wide basis.

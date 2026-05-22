<div align="center">

  <img src="./assets/icon.svg" alt="NordVPN icon" width="100" />

  <h1>NordVPN for Linux</h1>

  <h3>
    <strong>Privacy and security for Linux users</strong>
  </h3>

  <p>
    <a href="https://scorecard.dev/viewer/?uri=github.com/NordSecurity/nordvpn-linux">
      <img src="https://api.scorecard.dev/projects/github.com/NordSecurity/nordvpn-linux/badge" alt="OpenSSF Scorecard" />
    </a>
  </p>

  <h3>
    <a href="#about">About</a>
    <span> | </span>
    <a href="#versioning">Versioning</a>
    <span> | </span>
    <a href="#contributing">Contributing</a>
    <span> | </span>
    <a href="#building">Building</a>
    <span> | </span>
    <a href="#troubleshooting">Troubleshooting</a>
    <span> | </span>
    <a href="#installing">Installing</a>
  </h3>

</div>

# <p id="about">About</p>

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
  [nftables](https://nftables.org),
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

### Meshnet peer routing not working on Fedora with Docker installed

When Docker is installed on Fedora, it drops all forwarded traffic that did not
originate from Docker. This means meshnet routing through a Fedora machine will
not work.

To fix this, create `/etc/docker/daemon.json` with the following content:

```json
{
  "ip-forward-no-drop": true
}
```

Then restart the Docker socket service:

```sh
sudo systemctl restart docker.socket
```

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

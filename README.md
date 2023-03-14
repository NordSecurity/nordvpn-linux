# NordVPN for Linux

![icon](./assets/icon.svg)

---

[NordVPN](https://nordvpn.com/features/) The NordVPN Linux application provides a simple and user-friendly command line interface for connecting to NordVPN's servers.
Users can choose from a list of server locations around the world, or let the application automatically select the best server for them. 
They can also customize their connection settings, such as choosing a specific protocol or enabling the kill switch feature.


The application manages:
- network interfaces with the help of [tuntap](https://elixir.bootlin.com/linux/v6.0/source/Documentation/networking/tuntap.rst) kernel interface,
- firewall with the help of [iptables](https://www.netfilter.org/projects/iptables/index.html),
- routing with the help of [iproute2](https://wiki.linuxfoundation.org/networking/iproute2) and
- DNS with the help of [systemd-resolved](https://www.freedesktop.org/software/systemd/man/systemd-resolved.service.html).

---

# Versioning
The project follows https://semver.org/. Version tags and release branches must be named accordingly.

## Release tag
The correct way of releasing a new version is to create a version tag on `main`, named `X.X.X`.

# Contributing
We are happy to accept contibutions for the project. Please check out [Contribute.md](./CONTRIBUTE.md) file for more details on how to do so.

# Building
You can find everything related to building, testing and environment setup in [BUILD.md](BUILD.md).

# [Installing](https://nordvpn.com/download/linux/#install-nordvpn)

## Supported distros
* Ubuntu
* Fedora
* Manjaro
* Debian
* Arch
* Kali
* CentOS
* Rasbian

Distributions are not supported after their end of life.

This project is licensed under the terms of the GNU General Public License v3.0 only.
The registered trademark Linux® is used pursuant to a sublicense from the Linux Foundation, the exclusive licensee of Linus Torvalds, owner of the mark on a world-wide basis.

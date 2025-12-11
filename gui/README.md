# NordVPN Linux GUI application

## Overview

The [NordVPN](https://nordvpn.com/features/) Linux GUI application provides a
graphical user interface for accessing all the different features of NordVPN.
Users can choose from a list of server locations around the world, or let the
application automatically select the best server for them.
They can also customize their connection settings, such as choosing a specific
protocol or enabling the kill switch feature.

## Versioning

The project follows <https://semver.org/>. Version tags and release branches
must be named accordingly.

## Contributing

We are happy to accept contributions for the project. Please check out
[CONTRIBUTE.md](./CONTRIBUTE.md) file for more details on how to do so.

## Building

You can find everything related to building, testing and environment setup in [BUILD.md](./BUILD.md).

## Installing

For installing an already released version please follow the instructions on
our [official page](https://nordvpn.com/download/linux/#install-nordvpn).

## Debugging

Application stores logs in:

- `~/.cache/nordvpn/nordvpn-gui.log` for deb/rpm package
- `~/snap/nordvpn/current/.cache/nordvpn/nordvpn-gui.log` for snap package

By default it logs only `INFO` level. This can be controlled by setting
`NORDVPN_GUI_LOG_LEVEL` environment variable:

- for deb/rpm package:

```bash
NORDVPN_GUI_LOG_LEVEL=<level> nordvpn-gui
```

- for snap package:

```bash
NORDVPN_GUI_LOG_LEVEL=<level> nordvpn.nordvpn-gui
```

Supported levels (case insensitive):

- all
- trace
- debug
- info
- warn
- error
- fatal

## Supported distros

- Ubuntu
- Fedora
- Debian
- Kali
- OpenSUSE
- Raspbian

Distributions are not supported after their end of life.

This project is licensed under the terms of the
[GNU General Public License v3.0](../LICENSE.md) only.
The registered trademark Linux® is used pursuant to a sublicense from the
Linux Foundation, the exclusive licensee of Linus Torvalds, owner of the mark
on a world-wide basis.

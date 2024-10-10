# QA tests
Test application by imitating user actions.

## Overview
NordVPN application QA tests are written in Python using these 3rd packages for the following reasons:
* pytest - allows writing tests as functions and running code before/after each/all tests.
* requests - human friendly http client.
* sh - scans the `$PATH` on runtime and exposes found binaries as native Python functions.

Tests use pip as a package manager of choice. There are other package managers such as pipenv and poetry,
but only pip works all the time.

Test files must begin with `test_*.py` and each test must start with `def test_` or else pytest will not be able to run them.

## Package manager
By default, pip installs packages system-wide, which makes it really easy to break you code with OS updates.
This is what virtualenv is used for. In essence, virtualenv is a set of scripts and a copy of a Python interpreter.
Scripts are used to modify `$PATH`, so that packages are not installed/looked up globally anymore.

All you need to know about package manager to be able to run these tests is:
1. How to create virtualenv. 
```bash
python3 -m venv <name-of-the-directory> # which is .venv by convention and is already included in .gitignore
```
2. How to activate virtualenv.
```bash
source .venv/bin/activate # this should modify your shell prompt
```
3. How to install Python dependencies from file.
```bash
python3 -m pip install -r <path-to-requirements-txt> # requirements.txt is found in `ci/docker/tester/requirements.txt`
```
4. How to install new packages.
```bash
python3 -m pip install <package-name>
```
5. How to include your installed packages into requirements.txt
```bash
python3 -m pip freeze > <path-to-requirements-txt> # this also means that you will have to build and push a new qa docker image
```

## Setup
There are three ways to run QA tests:
* In CI
* Locally in docker
* Locally on the host

### CI
Running tests in CI is a simple button clicking.

### Docker
Running tests in docker requires Docker to be installed on the system. We have dedicated mage targets for testing.

In order to run QA tests, run `test:qaDocker` or `test:qaDockerFast`. It requires two arguments, name of the category and pattern for test functions to run.
Name of the test suite would be the name of the desired test file with *test* omitted. For test name, simply provide the desired function name from that file. Test names are selected by pattern matching. Since all test function names start with `test_`, simply use `test` as an argument to run every test in the suite.

For example, to run every test from the `test_fileshare.py` suite:

`mage test:qaDocker fileshare test`

And to run a single test:

`mage test:qaDocker fileshare test_accept`

To run tests without rebuilding everything each time use `test:qaDockerFast` instead of `test:qaDocker`.

Once tests are finished, logs are available at `dist/logs/daemon.log`.

### Locally
Running tests locally involves virtualenv creation, dependency installation etc.
For more details, take a look at `ci/test_deb.sh` script, because it wraps all commands needed to run tests.
To run only a specific test, use pytest's `-k` key which accepts a regexp.

## Test Categories
Tests are grouped in categories by how easy it is to test them in bulk and how they are related to each other.
When all tests in category require nearly identical setup/teardown, the code becomes a lot simpler.

Each category name is made by stripping `test_` and `.py` from the file name:
* autoconnect - tests autoconnect scenarios.
* combinations - tests every single reconnect combination.
* connect - tests IPv4 connections.
* connect6 - tests IPv6 connections.
* dns - tests DNS changes over the lifetime of IPv4 connection.
* dns6 - tests DNS changes over the lifetime of IPv6 connection.
* firewall - tests iptables rules.
* firewall6 - tests ip6tables rules.
* login - tests login scenarios.
* misc - tests socket permissions and domain rotations.
* allowlist - tests iptables and routing rules.

## Lib
Each Python file is treated as a module and directory with `__init__.py` in it is treated as a package
which contains modules.

lib is a Python package developed for making testing NordVPN easier:
* lib/daemon.py - allows easier control of the NordVPN daemon regardless of the init system.
* lib/dns.py - defines utility functions for checking dns during tests.
* lib/login.py - defines test user credentials in a single place.
* lib/network.py - allows easier control of the network.
* lib/server.py - helpful for calling core API.
* lib/settings.py - defines functions for checking application settings.

Package is imported by specifying a directory name (imports only `__init__.py`).
Module is imported by specifying path to file in the following format: `package.file`, where file
is a Python file without `.py`.

## Writing
Tests eventually fail and it's really hard to debug why if the failures are not reproducible.
Most of the time the root cause lies in shared resources between tests, which are not cleaned
up on failures. To make this easier, 2 classes where introduced: Defer and ErrorDefer.
- Defer is similar to Go's defer, but it works with blocks as well, not only function frames.
- ErrorDefer is taken from Zig and it is also similar Go's defer, but it is only executed in
case of the exception was thrown while inside the block.

Tests are written using assertions, which means that they are the culprits behind early exits.
If you see any asserts in the test cases, make sure they are called within Defer or ErrorDefer
blocks.

## Debugging
Sometimes using debugger is inevitable. In Python this is done by adding the following line in the code and running it:
```python
import pdb; pdb.set_trace()
```
The code will stop a line after the debugger statement and you will be given a debugger shell.

## REPL
Since Python is a dynamic language, it features REPL (Read, Evaluate, Print, Loop). REPL is similar to
system's shell such as bash, the main difference is that it runs only Python.
An example of REPL session.
```python
>>> network.stop()
>>> sudo.ip.addr()
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
222: eth0@if223: <BROADCAST,MULTICAST> mtu 1500 qdisc noqueue state DOWN group default 
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
>>> sh.nordvpn.account()
You are not logged in.
```

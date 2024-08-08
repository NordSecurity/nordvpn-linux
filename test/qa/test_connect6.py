import random

import pytest
import sh

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
from test_connect import disconnect_base_test


def setup_function(function):  # noqa: ARG001
    daemon.start()
    login.login_as("default")
    logging.log()
    print(sh.nordvpn.set.ipv6.on())


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    print(sh.nordvpn.set.ipv6.off())
    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


def connect_base_test(group: str = (), name: str = "", hostname: str = "", ipv6 = True):
    """
    Connects to a NordVPN server and performs a series of checks to ensure the connection is successful.

    Parameters:
    group (str): The specific server name or group name to connect to. Default is an empty string.
    name (str): Used to verify the connection message. Default is an empty string.
    hostname (str): Used to verify the connection message. Default is an empty string.
    ipv6 (bool): If True, checks if IPv6 connection is available. Default is True.
    """

    output = sh.nordvpn.connect(group, _tty_out=False)
    print(output)

    assert lib.is_connect_successful(output, name, hostname)
    assert network.is_connected()

    if ipv6:
        assert network.is_ipv6_connected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_WITH_IPV6)
def test_ipv6_connect(tech, proto, obfuscated) -> None:
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(random.choice(lib.IPV6_SERVERS))
    disconnect_base_test()


def test_ipv6_enabled_ipv4_connect():
    lib.set_technology_and_protocol(*lib.STANDARD_TECHNOLOGIES[0])
    connect_base_test("pl128", "Poland #128", "pl128.nordvpn.com", False)

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    disconnect_base_test()


def test_ipv6_double_connect_without_disconnect():
    lib.set_technology_and_protocol(*lib.STANDARD_TECHNOLOGIES[0])
    connect_base_test("pl128", "Poland #128", "pl128.nordvpn.com", False)

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    connect_base_test(random.choice(lib.IPV6_SERVERS))
    disconnect_base_test()

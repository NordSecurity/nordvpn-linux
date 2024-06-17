import random

import pytest
import sh
import timeout_decorator

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
from test_connect import connect_base_test, disconnect_base_test


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


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_WITH_IPV6)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_connect(tech, proto, obfuscated) -> None:
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), random.choice(lib.IPV6_SERVERS), ipv6 = True)
    disconnect_base_test()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_enabled_ipv4_connect():
    lib.set_technology_and_protocol(*lib.STANDARD_TECHNOLOGIES[0])
    connect_base_test(lib.STANDARD_TECHNOLOGIES[0], "pl128")

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    disconnect_base_test()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_double_connect_without_disconnect():
    lib.set_technology_and_protocol(*lib.STANDARD_TECHNOLOGIES[0])
    connect_base_test(lib.STANDARD_TECHNOLOGIES[0], "pl128")

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    connect_base_test(lib.STANDARD_TECHNOLOGIES[0], random.choice(lib.IPV6_SERVERS), ipv6 = True)
    disconnect_base_test()

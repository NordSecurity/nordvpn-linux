from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import random
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_WITH_IPV6[:-1])
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_connect(tech, proto, obfuscated):
    output = sh.nordvpn.set.ipv6.on()
    print(output)
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    output = sh.nordvpn.disconnect()
    print(output)
    sh.nordvpn.set.ipv6.off()
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_enabled_ipv4_connect():
    output = sh.nordvpn.set.ipv6.on()
    print(output)
    lib.set_technology_and_protocol("openvpn", "udp", "off")

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect("pl128")
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_connected()

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    output = sh.nordvpn.disconnect()
    print(output)
    sh.nordvpn.set.ipv6.off()
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_ipv6_double_connect_without_disconnect():
    output = sh.nordvpn.set.ipv6.on()
    print(output)
    lib.set_technology_and_protocol("openvpn", "udp", "off")

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect("pl128")
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_connected()

    with pytest.raises(sh.ErrorReturnCode_2) as ex:
        network.is_ipv6_connected(2)

    assert "Cannot assign requested address" in str(ex.value)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))
        print(output)
        assert lib.is_connect_successful(output)
        assert network.is_ipv4_and_ipv6_connected(20)

    output = sh.nordvpn.disconnect()
    print(output)
    sh.nordvpn.set.ipv6.off()
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

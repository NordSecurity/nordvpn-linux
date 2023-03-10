from lib import (
    daemon,
    dns,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
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


@pytest.mark.parametrize("threat_protection_lite", lib.THREAT_PROTECTION_LITE)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_dns_connect(tech, proto, obfuscated, threat_protection_lite):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_threat_protection_lite(threat_protection_lite)

    assert dns.is_unset()

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()
        assert dns.is_set_for(threat_protection_lite)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", lib.DNS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_custom_dns_connect(tech, proto, obfuscated, nameserver):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set.dns(nameserver)

    assert dns.is_unset()

    output = sh.nordvpn.connect()
    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()
        assert dns.is_unset()

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()

    sh.nordvpn.set.dns.off()

    assert dns.is_unset()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(20)
def test_set_dns_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_threat_protection_lite("off")

    assert dns.is_unset()

    output = sh.nordvpn.connect()

    print(output)
    assert lib.is_connect_successful(output)

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()
        assert dns.is_set_for("off")

    lib.set_threat_protection_lite("on")

    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()
        assert dns.is_set_for("on")

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert dns.is_unset()
    assert network.is_disconnected()

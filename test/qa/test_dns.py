from lib import (
    daemon,
    dns,
    info,
    logging,
    login,
    settings
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

    # Make sure that Custom DNS, IPv6 and Threat Protection Lite are disabled before we execute each test
    lib.set_dns("off")
    lib.set_ipv6("off")
    lib.set_threat_protection_lite("off")


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("threat_protection_lite", lib.THREAT_PROTECTION_LITE)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_dns_connect(tech, proto, obfuscated, threat_protection_lite):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_threat_protection_lite(threat_protection_lite)

    # Make sure, that DNS is unset before we connect to VPN server
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        if threat_protection_lite == "on":
            assert settings.get_is_tpl_enabled()
            assert settings.dns_visible_in_settings("disabled")
            assert dns.is_set_for(dns.DNS_TPL)
        else:
            assert not settings.get_is_tpl_enabled()
            assert settings.dns_visible_in_settings("disabled")
            assert dns.is_set_for(dns.DNS_NORD)

    # Make sure, that DNS is unset, after we disconnect from VPN server
    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_custom_dns_connect(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set.dns(nameserver)

    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not settings.get_is_tpl_enabled()
        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

    assert dns.is_unset()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_set_dns_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        assert not settings.get_is_tpl_enabled()
        assert settings.dns_visible_in_settings("disabled")
        assert dns.is_set_for(dns.DNS_NORD)

        lib.set_threat_protection_lite("on")
        assert settings.get_is_tpl_enabled()
        assert settings.dns_visible_in_settings("disabled")
        assert dns.is_set_for(dns.DNS_TPL)

    assert dns.is_unset()
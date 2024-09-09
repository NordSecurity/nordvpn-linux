import random

import pytest
import sh

import lib
from lib import daemon, dns, info, logging, login, settings


def setup_module(module):  # noqa: ARG001
    daemon.start()
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()

    # Make sure that Custom DNS, IPv6 and Threat Protection Lite are disabled before we execute each test
    lib.set_dns("off")
    lib.set_ipv6("off")
    lib.set_threat_protection_lite("off")


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("threat_protection_lite", lib.THREAT_PROTECTION_LITE)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_dns_connect(tech, proto, obfuscated, threat_protection_lite):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    lib.set_threat_protection_lite(threat_protection_lite)
    lib.set_ipv6("on")

    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))

        if threat_protection_lite == "on":
            assert settings.is_tpl_enabled()
            assert settings.dns_visible_in_settings(["disabled"])
            assert dns.is_set_for(dns.DNS_TPL_IPV6 + dns.DNS_TPL)
        else:
            assert not settings.is_tpl_enabled()
            assert settings.dns_visible_in_settings(["disabled"])
            assert dns.is_set_for(dns.DNS_NORD_IPV6 + dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_set_dns_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    # TODO: LVPN-1349
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()

    lib.set_ipv6("on")

    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect(random.choice(lib.IPV6_SERVERS))

        assert not settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_NORD_IPV6 + dns.DNS_NORD)

        lib.set_threat_protection_lite("on")
        assert settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_TPL_IPV6 + dns.DNS_TPL)

    assert dns.is_unset()

import pytest
import sh
import dns.resolver as dnspy

import lib
from lib import dns, settings


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs", "disable_dns_and_threat_protection")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_on_off_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    # Make sure, that DNS is unset before we connect to VPN server
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        tpl_alias = dns.get_tpl_alias()
        assert "Threat Protection Lite is set to 'enabled' successfully." in sh.nordvpn.set(tpl_alias, "on")

        assert settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_TPL)

        tpl_alias = dns.get_tpl_alias()
        assert "Threat Protection Lite is set to 'disabled' successfully." in sh.nordvpn.set(tpl_alias, "off")

        assert not settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_NORD)

    # Make sure, that DNS is unset, after we disconnect from VPN server
    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_on_and_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    assert "Threat Protection Lite is set to 'enabled' successfully." in sh.nordvpn.set(tpl_alias, "on")

    assert settings.is_tpl_enabled()
    assert settings.dns_visible_in_settings(["disabled"])
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert dns.is_set_for(dns.DNS_TPL)

    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_off_and_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    sh.nordvpn.set(tpl_alias, "on")

    assert "Threat Protection Lite is set to 'disabled' successfully." in sh.nordvpn.set(tpl_alias, "off")

    assert not settings.is_tpl_enabled()
    assert settings.dns_visible_in_settings(["disabled"])
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert dns.is_set_for(dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_tpl_on_set_custom_dns_disconnected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    sh.nordvpn.set(tpl_alias, "on")
    assert settings.is_tpl_enabled()

    output = sh.nordvpn.set.dns(nameserver)

    assert dns.TPL_MSG_WARNING_DISABLING in output
    assert not settings.is_tpl_enabled()
    assert settings.dns_visible_in_settings(nameserver)
    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_tpl_on_set_custom_dns_connected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        tpl_alias = dns.get_tpl_alias()
        sh.nordvpn.set(tpl_alias, "on")
        assert settings.is_tpl_enabled()

        output = sh.nordvpn.set.dns(nameserver)
        assert dns.TPL_MSG_WARNING_DISABLING in output
        assert not settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_connect(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set.dns(nameserver)

    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not settings.is_tpl_enabled()
        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_off_connect(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert settings.dns_visible_in_settings(nameserver)
    assert dns.is_unset()

    sh.nordvpn.set.dns("off")
    assert settings.dns_visible_in_settings(["disabled"])
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        assert dns.is_set_for(dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_custom_dns_connected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        sh.nordvpn.set.dns(nameserver)

        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_custom_dns_off_connected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        sh.nordvpn.set.dns(nameserver)
        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

        sh.nordvpn.set.dns("off")
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize(("nameserver", "expected_error"), dns.DNS_CASES_ERROR)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_errors_disconnected(tech, proto, obfuscated, nameserver, expected_error):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns(nameserver)

    assert expected_error in str(ex)
    assert settings.dns_visible_in_settings(["disabled"])
    assert dns.is_unset()


@pytest.mark.parametrize(("nameserver", "expected_error"), dns.DNS_CASES_ERROR)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_errors_connected(tech, proto, obfuscated, nameserver, expected_error):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns(nameserver)

        assert expected_error in str(ex)
        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.is_set_for(dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_set_disconnected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert dns.is_unset()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns(nameserver)

    full_error_message = dns.DNS_MSG_ERROR_ALREADY_SET % ", ".join(nameserver)

    assert full_error_message in str(ex)
    assert settings.dns_visible_in_settings(nameserver)
    assert dns.is_unset()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_set_connected(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert dns.is_unset()

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns(nameserver)

        full_error_message = dns.DNS_MSG_ERROR_ALREADY_SET % ", ".join(nameserver)
        assert full_error_message in str(ex)
        assert settings.dns_visible_in_settings(nameserver)
        assert dns.is_set_for(nameserver)

    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_disabled_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns("off")

    assert dns.DNS_MSG_ERROR_ALREADY_DISABLED in str(ex)
    assert settings.dns_visible_in_settings(["disabled"])
    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_disabled_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns("off")

        assert settings.dns_visible_in_settings(["disabled"])
        assert dns.DNS_MSG_ERROR_ALREADY_DISABLED in str(ex)
        assert dns.is_set_for(dns.DNS_NORD)

    assert dns.is_unset()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_order_is_kept(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    nameserver_list = ["8.8.8.8", "1.1.1.1"]
    sh.nordvpn.set.dns(nameserver_list)
    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        resolver = dnspy.Resolver()
        if "127.0.0.53" in resolver.nameservers:
            found = False
            if tech == "nordlynx":
                output = sh.resolvectl.status.nordlynx()
            if tech == "openvpn":
                output = sh.resolvectl.status.nordtun()
            for line in output:
                print(line)
                if 'DNS Servers' in line:
                    found = True
                    servers_str = line.split(": ")[1].rstrip("\n")
                    servers = servers_str.split(" ")
                    assert servers == nameserver_list
            assert found, "Nordlynx/Nordtun device or their DNS servers were not found"
        else:
            assert nameserver_list == resolver.nameservers
    assert dns.is_unset()

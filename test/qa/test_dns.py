import pytest
import sh
import dns.resolver as dnspy

import lib
from lib import dns, settings, IS_NIGHTLY
from lib.dynamic_parametrize import dynamic_parametrize

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs", "disable_dns_and_threat_protection")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_on_off_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-8718"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    # Make sure, that DNS is unset before we connect to VPN server
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        tpl_alias = dns.get_tpl_alias()
        assert "Threat Protection Lite has been successfully set to 'enabled'" in sh.nordvpn.set(tpl_alias, "on"), "TPL enable should show success message"

        assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on"
        assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings when TPL is enabled"
        assert dns.is_set_for(dns.DNS_TPL), "DNS should be set for TPL when connected with TPL enabled"

        tpl_alias = dns.get_tpl_alias()
        assert "Threat Protection Lite has been successfully set to 'disabled'." in sh.nordvpn.set(tpl_alias, "off"), "TPL disable should show success message"

        assert not settings.is_tpl_enabled(), "TPL should be disabled after setting it to off"
        assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after TPL is disabled"
        assert dns.is_set_for(dns.DNS_NORD), "DNS should be set for Nord DNS when TPL is disabled"

    # Make sure, that DNS is unset, after we disconnect from VPN server
    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_on_and_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-1603"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    assert "Threat Protection Lite has been successfully set to 'enabled'." in sh.nordvpn.set(tpl_alias, "on"), "TPL enable should show success message"

    assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on"
    assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings when TPL is enabled"
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert dns.is_set_for(dns.DNS_TPL), "DNS should be set for TPL when connected with TPL enabled"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_tpl_off_and_connect(tech, proto, obfuscated):
    """Manual TC: LVPN-1606"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    sh.nordvpn.set(tpl_alias, "on")

    assert "Threat Protection Lite has been successfully set to 'disabled'." in sh.nordvpn.set(tpl_alias, "off"), "TPL disable should show success message"

    assert not settings.is_tpl_enabled(), "TPL should be disabled after setting it to off"
    assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after TPL is disabled"
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert dns.is_set_for(dns.DNS_NORD), "DNS should be set for Nord DNS when connected with TPL disabled"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_tpl_on_set_custom_dns_disconnected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-6803"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    tpl_alias = dns.get_tpl_alias()
    sh.nordvpn.set(tpl_alias, "on")
    assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on"

    output = sh.nordvpn.set.dns(nameserver)

    assert dns.TPL_MSG_WARNING_DISABLING in output, "TPL warning message should appear when setting custom DNS with TPL enabled"
    assert not settings.is_tpl_enabled(), "TPL should be disabled when custom DNS is set"
    assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings"
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_tpl_on_set_custom_dns_connected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-6802"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        tpl_alias = dns.get_tpl_alias()
        sh.nordvpn.set(tpl_alias, "on")
        assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on while connected"

        output = sh.nordvpn.set.dns(nameserver)
        assert dns.TPL_MSG_WARNING_DISABLING in output, "TPL warning message should appear when setting custom DNS with TPL enabled while connected"
        assert not settings.is_tpl_enabled(), "TPL should be disabled when custom DNS is set while connected"
        assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings while connected"
        assert dns.is_set_for(nameserver), "DNS should be set for custom nameserver when connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_connect(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-6793"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set.dns(nameserver)

    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not settings.is_tpl_enabled(), "TPL should not be enabled after setting custom DNS"
        assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings when connected"
        assert dns.is_set_for(nameserver), "DNS should be set for custom nameserver when connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_off_connect(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-6796"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings"
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    sh.nordvpn.set.dns("off")
    assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after setting to off"
    assert dns.is_unset(), "DNS should still be unset after setting to off"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        assert dns.is_set_for(dns.DNS_NORD), "DNS should be set for Nord DNS when connected after disabling custom DNS"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_set_custom_dns_connected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-6790"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        sh.nordvpn.set.dns(nameserver)

        assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings when connected"
        assert dns.is_set_for(nameserver), "DNS should be set for custom nameserver when connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_set_custom_dns_off_connected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-1637"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        sh.nordvpn.set.dns(nameserver)
        assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings when connected"
        assert dns.is_set_for(nameserver), "DNS should be set for custom nameserver when connected"

        sh.nordvpn.set.dns("off")
        assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after setting to off while connected"
        assert dns.is_set_for(dns.DNS_NORD), "DNS should be set for Nord DNS when set to off while connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@pytest.mark.parametrize(("nameserver", "expected_error"), dns.DNS_CASES_ERROR)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_errors_disconnected(tech, proto, obfuscated, nameserver, expected_error):
    """Manual TC: LVPN-6799"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns(nameserver)

    assert expected_error in ex.value.stdout.decode("utf-8"), "Expected DNS error message should be present"
    assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after error"
    assert dns.is_unset(), "DNS should be unset after DNS set error"


@pytest.mark.parametrize(("nameserver", "expected_error"), dns.DNS_CASES_ERROR)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_errors_connected(tech, proto, obfuscated, nameserver, expected_error):
    """Manual TC: LVPN-6798"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns(nameserver)

        assert expected_error in ex.value.stdout.decode("utf-8"), "Expected DNS error message should be present when connected"
        assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after error while connected"
        assert dns.is_set_for(dns.DNS_NORD), "DNS should still be set for Nord DNS after DNS set error while connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_already_set_disconnected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8755"""

    nameserver = nameserver.split(" ")
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns(nameserver)

    full_error_message = dns.DNS_MSG_ERROR_ALREADY_SET % ", ".join(nameserver)

    assert full_error_message in ex.value.stdout.decode("utf-8"), "Error message should indicate DNS is already set"
    assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should remain visible in settings"
    assert dns.is_unset(), "DNS should still be unset after failed set attempt"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_already_set_connected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8754"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns(nameserver)

        full_error_message = dns.DNS_MSG_ERROR_ALREADY_SET % ", ".join(nameserver)
        assert full_error_message in ex.value.stdout.decode("utf-8"), "Error message should indicate DNS is already set while connected"
        assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should remain visible in settings while connected"
        assert dns.is_set_for(nameserver), "DNS should still be set for custom nameserver after failed set attempt"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_disabled_disconnected(tech, proto, obfuscated):
    """Manual TC: LVPN-8757"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.dns("off")

    assert dns.DNS_MSG_ERROR_ALREADY_DISABLED in ex.value.stdout.decode("utf-8"), "Error message should indicate DNS is already disabled"
    assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after error"
    assert dns.is_unset(), "DNS should be unset after DNS disable error"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_already_disabled_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-8756"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.dns("off")

        assert settings.dns_visible_in_settings(["disabled"]), "DNS should show as disabled in settings after error"
        assert dns.DNS_MSG_ERROR_ALREADY_DISABLED in ex.value.stdout.decode("utf-8"), "Error message should indicate DNS is already disabled while connected"
        assert dns.is_set_for(dns.DNS_NORD), "DNS should still be set for Nord DNS after DNS disable error while connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_custom_dns_order_is_kept(tech, proto, obfuscated):
    """Manual TC is unavailable because resolver settings and DNS order cannot be reliably verified without automation."""

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
                    assert servers == nameserver_list, "DNS server order should be preserved as configured"
            assert found, "Nordlynx/Nordtun device or their DNS servers were not found"
        else:
            assert nameserver_list == resolver.nameservers, "Resolver nameservers should match configured DNS order"
    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_removed_when_tpl_enabled_disconnected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8439"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    nameserver = nameserver.split(" ")

    sh.nordvpn.set.dns(nameserver)

    assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings"

    tpl_alias = dns.get_tpl_alias()
    output = sh.nordvpn.set(tpl_alias, "on")

    assert dns.DNS_MSG_WARNING_DISABLING in output, "TPL warning message should appear when enabling TPL with custom DNS set"
    assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on"
    assert not settings.dns_visible_in_settings(nameserver), "Custom nameserver should be removed from settings when TPL is enabled"
    assert dns.is_unset(), "DNS should be unset after TPL is enabled with custom DNS"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_custom_dns_removed_when_tpl_enabled_connected(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8444"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    nameserver = nameserver.split(" ")

    sh.nordvpn.set.dns(nameserver)

    assert settings.dns_visible_in_settings(nameserver), "Custom nameserver should be visible in settings"
    assert dns.is_unset(), "DNS should be unset before connecting to VPN server"

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert dns.is_set_for(nameserver), "DNS should be set for custom nameserver when connected"

        tpl_alias = dns.get_tpl_alias()
        output = sh.nordvpn.set(tpl_alias, "on")

        assert dns.DNS_MSG_WARNING_DISABLING in output, "TPL warning message should appear when enabling TPL with custom DNS set while connected"
        assert settings.is_tpl_enabled(), "TPL should be enabled after setting it to on while connected"
        assert not settings.dns_visible_in_settings(nameserver), "Custom nameserver should be removed from settings when TPL is enabled while connected"
        assert dns.is_set_for(dns.DNS_TPL), "DNS should be set for TPL after enabling TPL while connected"

    assert dns.is_unset(), "DNS should be unset after disconnecting from VPN server"

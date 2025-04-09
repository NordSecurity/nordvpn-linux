import pytest
import sh

import lib
from lib import daemon, dns, info, logging, login, network, settings


def setup_module(module):  # noqa: ARG001
    daemon.start()


def teardown_module(module):  # noqa: ARG001
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()
    login.login_as("default")


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()
    sh.nordvpn.set.defaults()


MSG_SET_DEFAULTS = "Settings were successfully restored to defaults."


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_obfuscate_nonobfucated(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.obfuscate("on")
        assert "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation." in str(ex.value)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC2 + lib.TECHNOLOGIES_BASIC1)
def test_set_technology(tech, proto, obfuscated):  # noqa: ARG001

    if tech == "nordlynx":
        sh.nordvpn.set.technology("OPENVPN")

    tech_name =  lib.technology_to_upper_camel_case(tech)
    assert f"Technology is set to '{tech_name}' successfully." in sh.nordvpn.set.technology(tech)
    assert tech.upper() in sh.nordvpn.settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_protocol_in_settings(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert proto.upper() in sh.nordvpn.settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_technology_set_options(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    ovpn_list = "obfuscate" in sh.nordvpn.set() and "protocol" in sh.nordvpn.set()

    if tech == "openvpn":
        assert ovpn_list
    else:
        assert not ovpn_list


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_logged_in_1st_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set.ipv6("on")
    sh.nordvpn.set.notify("on")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled()
    assert not settings.is_routing_enabled()
    assert not settings.is_dns_disabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_ipv6_enabled()
    assert settings.is_notify_enabled()
    assert not settings.is_virtual_location_enabled()

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    assert MSG_SET_DEFAULTS in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_logged_out_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.autoconnect("on")
    sh.nordvpn.set.notify("on")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.ipv6("on")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled()
    assert not settings.is_routing_enabled()
    assert settings.is_autoconnect_enabled()
    assert settings.is_notify_enabled()
    assert not settings.is_dns_disabled()
    assert settings.is_ipv6_enabled()
    assert not settings.is_virtual_location_enabled()

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    sh.nordvpn.logout("--persist-token")

    assert MSG_SET_DEFAULTS in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_connected_1st_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set("lan-discovery", "on")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status()

    assert not settings.is_routing_enabled()
    assert not settings.is_dns_disabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_lan_discovery_enabled()
    assert not settings.is_virtual_location_enabled()

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    assert MSG_SET_DEFAULTS in sh.nordvpn.set.defaults()

    assert "Status: Disconnected" in sh.nordvpn.status()

    assert settings.app_has_defaults_settings()


@pytest.mark.skip(reason="LVPN-265")
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_is_killswitch_disabled_after_setting_defaults(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.killswitch("on")
    assert network.is_not_available(2)

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status()
    assert network.is_available()

    assert daemon.is_killswitch_on()

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    assert MSG_SET_DEFAULTS in sh.nordvpn.set.defaults()

    assert "Status: Disconnected" in sh.nordvpn.status()
    assert network.is_available()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize("nameserver", dns.DNS_CASES_CUSTOM)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_is_custom_dns_removed_after_setting_defaults(tech, proto, obfuscated, nameserver):
    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert settings.dns_visible_in_settings(nameserver)

    sh.nordvpn.connect()

    assert dns.is_set_for(nameserver)

    assert MSG_SET_DEFAULTS in sh.nordvpn.set.defaults()

    login.login_as("default")

    assert settings.app_has_defaults_settings()

    sh.nordvpn.connect()

    assert not dns.is_set_for(nameserver)


def test_set_analytics_off_on():

    assert "Analytics is set to 'disabled' successfully." in sh.nordvpn.set.analytics("off")
    assert not settings.are_analytics_enabled()

    assert "Analytics is set to 'enabled' successfully." in sh.nordvpn.set.analytics("on")
    assert settings.are_analytics_enabled()


def test_set_analytics_on_off_repeated():

    assert "Analytics is already set to 'enabled'." in sh.nordvpn.set.analytics("on")

    sh.nordvpn.set.analytics("off")
    assert "Analytics is already set to 'disabled'." in sh.nordvpn.set.analytics("off")


def test_set_virtual_location_off_on():

    assert "Virtual location is set to 'disabled' successfully." in sh.nordvpn.set("virtual-location", "off")
    assert not settings.is_virtual_location_enabled()

    assert "Virtual location is set to 'enabled' successfully." in sh.nordvpn.set("virtual-location", "on")
    assert settings.is_virtual_location_enabled()


def test_set_virtual_location_on_off_repeated():

    assert "Virtual location is already set to 'enabled'." in sh.nordvpn.set("virtual-location", "on")

    sh.nordvpn.set("virtual-location", "off")
    assert "Virtual location is already set to 'disabled'." in sh.nordvpn.set("virtual-location", "off")


def test_set_post_quantum_on_off():

    pq_alias = settings.get_pq_alias()

    assert "Post-quantum VPN is set to 'enabled' successfully." in sh.nordvpn.set(pq_alias, "on")
    assert not settings.is_post_quantum_disabled()

    assert "Post-quantum VPN is set to 'disabled' successfully." in sh.nordvpn.set(pq_alias, "off")
    assert settings.is_post_quantum_disabled()


def test_set_post_quantum_off_on_repeated():

    pq_alias = settings.get_pq_alias()

    assert "Post-quantum VPN is already set to 'disabled'." in sh.nordvpn.set(pq_alias, "off")

    sh.nordvpn.set(pq_alias, "on")
    assert "Post-quantum VPN is already set to 'enabled'." in sh.nordvpn.set(pq_alias, "on")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES + lib.OBFUSCATED_TECHNOLOGIES)
def test_set_post_quantum_on_open_vpn(tech, proto, obfuscated):

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption is unavailable with OpenVPN. Switch to NordLynx to activate post-quantum protection." in str(ex.value)


def test_set_technology_openvpn_post_quantum_enabled():

    sh.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.technology("OPENVPN")

    assert "This setting is not compatible with post-quantum encryption. To use OpenVPN, disable post-quantum encryption first." in str(ex.value)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_enable_twice(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for _ in range(2):
        output = sh.nordvpn.set.autoconnect.on()
        print(output)
        assert settings.MSG_AUTOCONNECT_ENABLE_SUCCESS in output


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_disable_twice(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(str(output))
    assert settings.MSG_AUTOCONNECT_DISABLE_FAIL in str(output)

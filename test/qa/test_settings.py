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


autoconnect_on_parameters = [
    ("lt16", "on", "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."),
    ("uk2188", "off", "Turn on obfuscation to connect to obfuscated servers.")
]


MSG_SET_DEFAULTS = "Settings were successfully restored to defaults."


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_obfuscate_nonobfucated(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.obfuscate("on")
        assert "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation." in str(ex.value)


@pytest.mark.skip(reason="LVPN-2119")
@pytest.mark.parametrize(("server", "obfuscated", "error_message"), autoconnect_on_parameters)
def test_autoconnect_on_server_obfuscation_mismatch(server, obfuscated, error_message):
    lib.set_technology_and_protocol("openvpn", "tcp", obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.autoconnect.on(server)

    print(ex.value)
    assert error_message in str(ex.value)

    assert "Auto-connect: disabled" in sh.nordvpn.settings()

    daemon.restart()
    assert network.is_disconnected()

    sh.nordvpn.set.autoconnect.off()


set_obfuscate_parameters = [
    ("off", "lt16", "We couldn’t turn on obfuscation because the current auto-connect server doesn’t support it. Set a different server for auto-connect to use obfuscation."),
    ("on", "uk2188", "We couldn’t turn off obfuscation because your current auto-connect server is obfuscated by default. Set a different server for auto-connect, then turn off obfuscation.")
]


@pytest.mark.skip(reason="LVPN-2119")
@pytest.mark.parametrize(("obfuscate_initial_state", "server", "error_message"), set_obfuscate_parameters)
def test_set_obfuscate_server_obfuscation_mismatch(obfuscate_initial_state, server, error_message):
    lib.set_technology_and_protocol("openvpn", "tcp", obfuscate_initial_state)

    output = sh.nordvpn.set.autoconnect.on(server)
    print(output)

    obfuscate_expected_state = "disabled"
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if obfuscate_initial_state == "off":
            sh.nordvpn.set.obfuscate.on()
        else:
            obfuscate_expected_state = "enabled"
            sh.nordvpn.set.obfuscate.off()

    assert f"Obfuscate: {obfuscate_expected_state}" in sh.nordvpn.settings()

    assert error_message in str(ex.value)

    sh.nordvpn.set.autoconnect.off()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC2 + lib.TECHNOLOGIES_BASIC1)
def test_set_technology(tech, proto, obfuscated):  # noqa: ARG001

    if tech == "nordlynx":
        sh.nordvpn.set.technology("OPENVPN")

    assert f"Technology is set to '{tech.upper()}' successfully." in sh.nordvpn.set.technology(tech)
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

    assert not settings.is_firewall_enabled()
    assert not settings.is_routing_enabled()
    assert not settings.is_dns_disabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_ipv6_enabled()
    assert settings.is_notify_enabled()
    assert not settings.is_virtual_location_enabled()

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

    assert not settings.is_firewall_enabled()
    assert not settings.is_routing_enabled()
    assert settings.is_autoconnect_enabled()
    assert settings.is_notify_enabled()
    assert not settings.is_dns_disabled()
    assert settings.is_ipv6_enabled()
    assert not settings.is_virtual_location_enabled()

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

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status()

    assert not settings.is_routing_enabled()
    assert not settings.is_dns_disabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_lan_discovery_enabled()
    assert not settings.is_virtual_location_enabled()

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

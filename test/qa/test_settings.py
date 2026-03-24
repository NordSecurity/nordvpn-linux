import pytest
import sh
import pexpect

import lib
from lib import daemon, dns, info, logging, login, network, settings, IS_NIGHTLY
from lib.dynamic_parametrize import dynamic_parametrize


def setup_function(function):  # noqa: ARG001
    logging.log()
    daemon.start()
    login.login_as("default")


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()
    sh.nordvpn.set.defaults("--logout")
    daemon.stop()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1 + lib.NORDWHISPER_TECHNOLOGY)
def test_obfuscate_nonobfucated(tech, proto, obfuscated):
    """Manual TC: LVPN-788"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available(), "Network should be available before attempting to set obfuscation"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.obfuscate("on")
        assert "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation." in ex.value.stdout.decode("utf-8")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC2 + lib.TECHNOLOGIES_BASIC1 + lib.NORDWHISPER_TECHNOLOGY)
def test_set_technology(tech, proto, obfuscated):  # noqa: ARG001
    """Manual TC: LVPN-601"""

    if tech == "nordlynx":
        sh.nordvpn.set.technology("OPENVPN")

    tech_name =  lib.technology_to_upper_camel_case(tech)
    assert f"Technology has been successfully set to '{tech_name}'." in sh.nordvpn.set.technology(tech), "Technology should be successfully set"
    assert tech.upper() in sh.nordvpn.settings(), "Technology should appear in settings"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_protocol_in_settings(tech, proto, obfuscated):
    """Manual TC: LVPN-8793"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert proto.upper() in sh.nordvpn.settings(), "Protocol should appear in settings"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_technology_set_options(tech, proto, obfuscated):
    """Manual TC: LVPN-6816"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    ovpn_list = "obfuscate" in sh.nordvpn.set() and "protocol" in sh.nordvpn.set()

    if tech == "openvpn":
        assert ovpn_list, "OpenVPN should have obfuscate and protocol options available"
    else:
        assert not ovpn_list, "Non-OpenVPN technology should not have obfuscate and protocol options"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_logged_in_1st_set(tech, proto, obfuscated):
    """Manual TC: LVPN-8737"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set.notify("on")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled(), "Firewall should be disabled"
    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"
    assert settings.is_user_consent_declared(), "User consent should be declared"
    assert settings.is_notify_enabled(), "Notifications should be enabled"
    assert not settings.is_virtual_location_enabled(), "Virtual location should be disabled"

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled(), "Obfuscation should be enabled"
    else:
        assert not settings.is_obfuscated_enabled(), "Obfuscation should be disabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_logged_out_2nd_set(tech, proto, obfuscated):
    """Manual TC: LVPN-8829"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.autoconnect("on")
    sh.nordvpn.set.notify("on")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled(), "Firewall should be disabled"
    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert settings.is_autoconnect_enabled(), "Autoconnect should be enabled"
    assert settings.is_notify_enabled(), "Notifications should be enabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"
    assert not settings.is_virtual_location_enabled(), "Virtual location should be disabled"

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled(), "Obfuscation should be enabled"
    else:
        assert not settings.is_obfuscated_enabled(), "Obfuscation should be disabled"

    sh.nordvpn.logout("--persist-token")

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_when_connected_1st_set(tech, proto, obfuscated):
    """Manual TC: LVPN-8741"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set("lan-discovery", "on")
    sh.nordvpn.set("virtual-location", "off")

    if tech == "nordlynx":
        sh.nordvpn.set.pq("on")

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"

    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"
    assert settings.is_user_consent_declared(), "User consent should be declared"
    assert settings.is_lan_discovery_enabled(), "LAN discovery should be enabled"
    assert not settings.is_virtual_location_enabled(), "Virtual location should be disabled"

    if tech == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled(), "Obfuscation should be enabled"
    else:
        assert not settings.is_obfuscated_enabled(), "Obfuscation should be disabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_is_killswitch_disabled_after_setting_defaults(tech, proto, obfuscated):
    """Manual TC: LVPN-8749"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.killswitch("on")
    assert network.is_not_available(2), "Network should not be available with killswitch enabled"

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"
    assert network.is_available(), "Network should be available when connected with killswitch enabled"

    assert daemon.is_killswitch_on(), "Killswitch should be enabled"

    if obfuscated == "on":
        assert settings.is_obfuscated_enabled(), "Obfuscation should be enabled"
    else:
        assert not settings.is_obfuscated_enabled(), "Obfuscation should be disabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout", "--off-killswitch"), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"
    assert network.is_available(), "Network should be available after turning off killswitch"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "nameserver",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{nameserver}",
)
def test_is_custom_dns_removed_after_setting_defaults(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8747"""

    nameserver = nameserver.split(" ")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns(nameserver)
    assert settings.dns_visible_in_settings(nameserver), "Custom DNS should be visible in settings"

    sh.nordvpn.connect()

    assert dns.is_set_for(nameserver), "Custom DNS should be set when connected"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    login.login_as("default")

    assert settings.app_has_defaults_settings(), "App should have default settings"

    sh.nordvpn.connect()

    assert not dns.is_set_for(nameserver), "Custom DNS should be removed after defaults reset"


def test_set_analytics_starts_prompt_even_if_completed_before():
    """Manual TC: LVPN-8473"""

    # first run: see prompt and respond
    cli1 = pexpect.spawn("nordvpn", args=["set", "analytics"], encoding='utf-8', timeout=10)
    cli1.expect(lib.USER_CONSENT_PROMPT)
    output1 = cli1.before + cli1.after

    assert (
        lib.squash_whitespace(lib.EXPECTED_CONSENT_MESSAGE)
        in lib.squash_whitespace(output1)
    ), "Consent message did not match expected full output on first run"

    cli1.sendline("n")
    cli1.expect(pexpect.EOF)

    # second run: should see the prompt again
    cli2 = pexpect.spawn("nordvpn", args=["set", "analytics"], encoding='utf-8', timeout=10)
    cli2.expect(lib.USER_CONSENT_PROMPT)
    output2 = cli2.before + cli2.after

    assert (
        lib.squash_whitespace(lib.EXPECTED_CONSENT_MESSAGE)
        in lib.squash_whitespace(output2)
    ), "Consent message did not appear again on second run"

    cli2.sendline("y")
    cli2.expect(pexpect.EOF)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_no_logout(tech, proto, obfuscated):
    """Manual TC: LVPN-9029"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set("virtual-location", "off")
    sh.nordvpn.set("lan-discovery", "on")

    assert not settings.is_virtual_location_enabled(), "Virtual location should be disabled"
    assert settings.is_lan_discovery_enabled(), "LAN discovery should be enabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults(), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"
    assert "Account information" in sh.nordvpn.account(), "Account information should be displayed"


def test_set_analytics_off_on():
    """Manual TC: LVPN-510"""

    assert "Analytics has been successfully set to 'disabled'." in sh.nordvpn.set.analytics("off"), "Analytics should be successfully disabled"
    assert not settings.is_user_consent_granted(), "User consent should not be granted when analytics is disabled"

    assert "Analytics has been successfully set to 'enabled'." in sh.nordvpn.set.analytics("on"), "Analytics should be successfully enabled"
    assert settings.is_user_consent_granted(), "User consent should be granted when analytics is enabled"


def test_set_analytics_on_off_repeated():
    """Manual TC: LVPN-509"""

    assert "Analytics is already set to 'enabled'." in sh.nordvpn.set.analytics("on"), "Analytics should be already enabled"

    sh.nordvpn.set.analytics("off")
    assert "Analytics is already set to 'disabled'." in sh.nordvpn.set.analytics("off"), "Analytics should be already disabled"


def test_set_virtual_location_off_on():
    """Manual TC: LVPN-5253"""

    assert "Virtual location has been successfully set to 'disabled'." in sh.nordvpn.set("virtual-location", "off"), "Virtual location should be successfully disabled"
    assert not settings.is_virtual_location_enabled(), "Virtual location should be disabled"

    assert "Virtual location has been successfully set to 'enabled'." in sh.nordvpn.set("virtual-location", "on"), "Virtual location should be successfully enabled"
    assert settings.is_virtual_location_enabled(), "Virtual location should be enabled"


def test_set_virtual_location_on_off_repeated():
    """Manual TC: LVPN-5254"""

    assert "Virtual location is already set to 'enabled'." in sh.nordvpn.set("virtual-location", "on"), "Virtual location should be already enabled"

    sh.nordvpn.set("virtual-location", "off")
    assert "Virtual location is already set to 'disabled'." in sh.nordvpn.set("virtual-location", "off"), "Virtual location should be already disabled"


def test_set_post_quantum_on_off():
    """Manual TC: LVPN-5774"""

    pq_alias = settings.get_pq_alias()

    assert "Post-quantum VPN has been successfully set to 'enabled'." in sh.nordvpn.set(pq_alias, "on"), "Post-quantum should be successfully enabled"
    assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled"

    assert "Post-quantum VPN has been successfully set to 'disabled'." in sh.nordvpn.set(pq_alias, "off"), "Post-quantum should be successfully disabled"
    assert settings.is_post_quantum_disabled(), "Post-quantum should be disabled"


def test_set_post_quantum_off_on_repeated():
    """Manual TC: LVPN-5774"""

    pq_alias = settings.get_pq_alias()

    assert "Post-quantum VPN is already set to 'disabled'." in sh.nordvpn.set(pq_alias, "off"), "Post-quantum should be already disabled"

    sh.nordvpn.set(pq_alias, "on")
    assert "Post-quantum VPN is already set to 'enabled'." in sh.nordvpn.set(pq_alias, "on"), "Post-quantum should be already enabled"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES + lib.OBFUSCATED_TECHNOLOGIES)
def test_set_post_quantum_on_open_vpn(tech, proto, obfuscated):
    """Manual TC: LVPN-5787"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption is not compatible with OpenVPN. Switch to NordLynx to use this encryption." in ex.value.stdout.decode("utf-8")

@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_set_post_quantum_on_nordwhisper(tech, proto, obfuscated):
    """Manual TC: LVPN-8445"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption is not compatible with NordWhisper. Switch to NordLynx to use this encryption." in ex.value.stdout.decode("utf-8")

def test_set_technology_openvpn_post_quantum_enabled():
    """Manual TC: LVPN-8536"""

    sh.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.technology("OPENVPN")

    assert "This setting is not compatible with post-quantum encryption. To use OpenVPN, turn off post-quantum encryption first." in ex.value.stdout.decode("utf-8")

def test_set_technology_nordwhisper_post_quantum_enabled():
    """Manual TC: LVPN-6835"""

    sh.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.technology("NORDWHISPER")

    assert "This setting is not compatible with post-quantum encryption. To use NordWhisper, turn off post-quantum encryption first." in ex.value.stdout.decode("utf-8")

@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_enable_twice(tech, proto, obfuscated):
    """Manual TC: LVPN-8597"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for _ in range(2):
        output = sh.nordvpn.set.autoconnect.on()
        print(output)
        assert settings.MSG_AUTOCONNECT_ENABLE_SUCCESS in output, "Autoconnect enable success message should be shown"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_disable_twice(tech, proto, obfuscated):
    """Manual TC: LVPN-8583"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(str(output))
    assert settings.MSG_AUTOCONNECT_DISABLE_FAIL in str(output), "Autoconnect disable failure message should be shown"


@pytest.mark.parametrize("killswitch_initial", [True, False])
@pytest.mark.parametrize("killswitch_flag", [True, False])
def test_set_defaults_killswitch_interaction(killswitch_initial, killswitch_flag):
    """Manual TC: LVPN-8750"""

    try:
        sh.nordvpn.set.killswitch(str(killswitch_initial))
    except sh.ErrorReturnCode_1 as ex:
        assert "Kill Switch is already set to" in ex.value.stdout.decode("utf-8"), "Unexpected error returned by 'set killswitch'. Expected 'Killswitch already set to enabled/disabled."

    if killswitch_flag:
        sh.nordvpn.set.defaults("--off-killswitch")
    else:
        sh.nordvpn.set.defaults()

    expected_killswitch_state = killswitch_initial and not killswitch_flag

    assert daemon.is_killswitch_on() is expected_killswitch_state, f"Killswitch state should be {expected_killswitch_state}"
    assert network.is_not_available(2) is expected_killswitch_state, f"Network availability should be {not expected_killswitch_state}"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1 + lib.NORDWHISPER_TECHNOLOGY)
def test_set_protocol_openvpn_only(tech, proto, obfuscated):
    """Manual TC: LVPN-8537"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.protocol("TCP")
        assert "This setting is only available when the selected protocol is OpenVPN." in ex.value.stdout.decode("utf-8")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_set_defaults_no_logout_connected(tech, proto, obfuscated):
    """Manual TC: LVPN-9014"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set("notify", "off")
    sh.nordvpn.set("tpl", "on")

    sh.nordvpn.connect()

    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"
    assert not settings.is_notify_enabled(), "Notifications should be disabled"
    assert settings.is_tpl_enabled(), "TPL should be enabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults(), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"
    assert settings.app_has_defaults_settings(), "App should have default settings"
    assert "Account information" in sh.nordvpn.account(), "Account information should be displayed"


@pytest.mark.parametrize("nameserver", (dns.DNS_CASE_CUSTOM_SINGLE,))
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_is_custom_dns_removed_after_setting_defaults_no_logout(tech, proto, obfuscated, nameserver):
    """Manual TC: LVPN-8748"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.dns([nameserver])
    assert settings.dns_visible_in_settings([nameserver]), "Custom DNS should be visible in settings"

    sh.nordvpn.connect()

    assert dns.is_set_for([nameserver]), "Custom DNS should be set when connected"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults(), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"

    sh.nordvpn.connect()

    assert not dns.is_set_for(nameserver), "Custom DNS should be removed after defaults reset"


def test_tray_off_on():
    """Manual TC: LVPN-8776"""

    assert "Tray set to 'disabled' successfully." in sh.nordvpn.set.tray("off"), "Tray should be successfully disabled"
    assert not settings.is_tray_enabled(), "Tray should be disabled"

    assert "Tray set to 'enabled' successfully." in sh.nordvpn.set.tray("on"), "Tray should be successfully enabled"
    assert settings.is_tray_enabled(), "Tray should be enabled"


def test_tray_on_off_repeated():
    """Manual TC: LVPN-8778"""

    assert "Tray is already set to 'enabled'." in sh.nordvpn.set.tray("on"), "Tray should be already enabled"

    sh.nordvpn.set.tray("off")

    assert "Tray is already set to 'disabled'." in sh.nordvpn.set.tray("off"), "Tray should be already disabled"


def test_lan_discovery_on_off():
    """Manual TC: LVPN-8448"""

    assert "LAN Discovery has been successfully set to 'enabled'." in sh.nordvpn.set("lan-discovery", "on"), "LAN Discovery should be successfully enabled"
    assert settings.is_lan_discovery_enabled(), "LAN Discovery should be enabled"

    assert "LAN Discovery has been successfully set to 'disabled'." in sh.nordvpn.set("lan-discovery", "off"), "LAN Discovery should be successfully disabled"
    assert not settings.is_lan_discovery_enabled(), "LAN Discovery should be disabled"


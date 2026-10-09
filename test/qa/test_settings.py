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


@pytest.mark.parametrize("vpn_protocol", lib.STANDARD_VPN_PROTOCOLS)
def test_set_vpn_protocol(vpn_protocol):  # noqa: ARG001
    """Manual TC: LVPN-601"""

    if vpn_protocol == "nordlynx":
        lib.set_vpn_protocol("openvpn_udp")

    name = lib.vpn_protocol_display_name(vpn_protocol)
    output = sh.nordvpn.set.protocol(vpn_protocol)
    assert f"VPN Protocol has been successfully set to '{name}'." in output, "VPN protocol should be successfully set"
    assert settings.Settings().get("Protocol") == name.lower(), "VPN protocol should appear in settings"


@pytest.mark.parametrize("vpn_protocol", lib.OVPN_VPN_PROTOCOLS)
def test_protocol_in_settings(vpn_protocol):
    """Manual TC: LVPN-601"""

    lib.set_vpn_protocol(vpn_protocol)
    assert lib.vpn_protocol_display_name(vpn_protocol) in sh.nordvpn.settings(), "Protocol should appear in settings"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_technology_set_options(vpn_protocol):
    """
    Manual TC: LVPN-601.

    Every technology offers `nordvpn set protocol`.
    """
    lib.set_vpn_protocol(vpn_protocol)

    offered = settings.get_set_subcommands()

    assert "protocol" in offered, f"'{vpn_protocol}' should offer 'nordvpn set protocol', got {offered}"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_set_defaults_when_logged_in_1st_set(vpn_protocol):
    """Manual TC: LVPN-8737"""

    lib.set_vpn_protocol(vpn_protocol)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set.notify("on")

    if vpn_protocol == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled(), "Firewall should be disabled"
    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"
    assert settings.is_user_consent_declared(), "User consent should be declared"
    assert settings.is_notify_enabled(), "Notifications should be enabled"

    if vpn_protocol == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_set_defaults_when_logged_out_2nd_set(vpn_protocol):
    """Manual TC: LVPN-8829"""

    lib.set_vpn_protocol(vpn_protocol)

    daemon.restart() # Temporary solution to avoid Firewall staying enabled in settings - LVPN-4121

    sh.nordvpn.set.firewall("off")
    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.autoconnect("on")
    sh.nordvpn.set.notify("on")
    sh.nordvpn.set.dns("1.1.1.1")

    if vpn_protocol == "nordlynx":
        sh.nordvpn.set.pq("on")

    assert not settings.is_firewall_enabled(), "Firewall should be disabled"
    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert settings.is_autoconnect_enabled(), "Autoconnect should be enabled"
    assert settings.is_notify_enabled(), "Notifications should be enabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"

    if vpn_protocol == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    sh.nordvpn.logout()

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_set_defaults_when_connected_1st_set(vpn_protocol):
    """Manual TC: LVPN-8741"""

    lib.set_vpn_protocol(vpn_protocol)

    sh.nordvpn.set.routing("off")
    sh.nordvpn.set.dns("1.1.1.1")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set("lan-discovery", "on")

    if vpn_protocol == "nordlynx":
        sh.nordvpn.set.pq("on")

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"

    assert not settings.is_routing_enabled(), "Routing should be disabled"
    assert not settings.is_dns_disabled(), "DNS should be enabled"
    assert settings.is_user_consent_declared(), "User consent should be declared"
    assert settings.is_lan_discovery_enabled(), "LAN discovery should be enabled"

    if vpn_protocol == "nordlynx":
        assert not settings.is_post_quantum_disabled(), "Post-quantum should be enabled for NordLynx"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout"), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_is_killswitch_disabled_after_setting_defaults(vpn_protocol):
    """Manual TC: LVPN-8749"""

    lib.set_vpn_protocol(vpn_protocol)

    sh.nordvpn.set.killswitch("on")
    assert network.is_not_available(2), "Network should not be available with killswitch enabled"

    sh.nordvpn.connect()
    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"
    assert network.is_available(), "Network should be available when connected with killswitch enabled"

    assert daemon.is_killswitch_on(), "Killswitch should be enabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults("--logout", "--off-killswitch"), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"
    assert network.is_available(), "Network should be available after turning off killswitch"

    assert settings.app_has_defaults_settings(), "App should have default settings"


@dynamic_parametrize(
    [
        "vpn_protocol", "nameserver",
    ],
    ordered_source=[lib.VPN_PROTOCOLS],
    randomized_source=[dns.DNS_CASES_CUSTOM],
    generate_all=IS_NIGHTLY,
    id_pattern="{vpn_protocol}-{nameserver}",
)
def test_is_custom_dns_removed_after_setting_defaults(vpn_protocol, nameserver):
    """Manual TC: LVPN-8747"""

    nameserver = nameserver.split(" ")

    lib.set_vpn_protocol(vpn_protocol)

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


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_set_defaults_no_logout(vpn_protocol):
    """Manual TC: LVPN-9029"""

    lib.set_vpn_protocol(vpn_protocol)

    sh.nordvpn.set("lan-discovery", "on")

    assert settings.is_lan_discovery_enabled(), "LAN discovery should be enabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults(), "Defaults reset message should be shown"

    assert settings.app_has_defaults_settings(), "App should have default settings"
    assert "Account information" in sh.nordvpn.account(), "Account information should be displayed"


def test_set_analytics_off_on():
    """Manual TC: LVPN-509"""

    assert "Analytics has been successfully set to 'disabled'." in sh.nordvpn.set.analytics("off"), "Analytics should be successfully disabled"
    assert not settings.is_user_consent_granted(), "User consent should not be granted when analytics is disabled"

    assert "Analytics has been successfully set to 'enabled'." in sh.nordvpn.set.analytics("on"), "Analytics should be successfully enabled"
    assert settings.is_user_consent_granted(), "User consent should be granted when analytics is enabled"


def test_set_analytics_on_off_repeated():
    """Manual TC: LVPN-509"""

    assert "Analytics is already set to 'enabled'." in sh.nordvpn.set.analytics("on"), "Analytics should be already enabled"

    sh.nordvpn.set.analytics("off")
    assert "Analytics is already set to 'disabled'." in sh.nordvpn.set.analytics("off"), "Analytics should be already disabled"


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


@pytest.mark.parametrize("vpn_protocol", lib.OVPN_VPN_PROTOCOLS)
def test_set_post_quantum_on_open_vpn(vpn_protocol):
    """Manual TC: LVPN-5787"""

    lib.set_vpn_protocol(vpn_protocol)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption is not compatible with OpenVPN. Switch to NordLynx to use this encryption." in ex.value.stdout.decode("utf-8")

@pytest.mark.parametrize("vpn_protocol", lib.NORDWHISPER_VPN_PROTOCOL)
def test_set_post_quantum_on_nordwhisper(vpn_protocol):
    """Manual TC: LVPN-5787"""

    lib.set_vpn_protocol(vpn_protocol)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set(settings.get_pq_alias(), "on")

    assert "Post-quantum encryption is not compatible with NordWhisper. Switch to NordLynx to use this encryption." in ex.value.stdout.decode("utf-8")

@pytest.mark.parametrize("vpn_protocol", lib.OVPN_VPN_PROTOCOLS)
def test_set_protocol_openvpn_post_quantum_enabled(vpn_protocol):  # noqa: ARG001
    """Manual TC: LVPN-6835"""

    sh.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.protocol(vpn_protocol)

    name = lib.vpn_protocol_display_name(vpn_protocol)
    assert f"This setting is not compatible with post-quantum encryption. To use {name}, turn off post-quantum encryption first." in ex.value.stdout.decode("utf-8")
    assert settings.Settings().get("Protocol") == "nordlynx", "VPN protocol should stay NordLynx"

def test_set_protocol_nordwhisper_post_quantum_enabled():
    """Manual TC: LVPN-6835"""

    sh.nordvpn.set(settings.get_pq_alias(), "on")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.protocol("nordwhisper")

    assert "This setting is not compatible with post-quantum encryption. To use NordWhisper, turn off post-quantum encryption first." in ex.value.stdout.decode("utf-8")
    assert settings.Settings().get("Protocol") == "nordlynx", "VPN protocol should stay NordLynx"

@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_autoconnect_enable_twice(vpn_protocol):
    """Manual TC: LVPN-8597"""

    lib.set_vpn_protocol(vpn_protocol)

    for _ in range(2):
        output = sh.nordvpn.set.autoconnect.on()
        print(output)
        assert settings.MSG_AUTOCONNECT_ENABLE_SUCCESS in output, "Autoconnect enable success message should be shown"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_autoconnect_disable_twice(vpn_protocol):
    """Manual TC: LVPN-8583"""

    lib.set_vpn_protocol(vpn_protocol)

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


@pytest.mark.parametrize("value", ["udp", "tcp", "openvpn", "unspecified"])
def test_set_protocol_rejects_invalid_value(value):
    """Manual TC: LVPN-8537"""

    protocol_before = settings.Settings().get("Protocol")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.protocol(value)

    assert "The command you entered is not valid." in ex.value.stdout.decode("utf-8")
    assert settings.Settings().get("Protocol") == protocol_before, "VPN protocol should not change"


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_set_defaults_no_logout_connected(vpn_protocol):
    """Manual TC: LVPN-9014"""

    lib.set_vpn_protocol(vpn_protocol)

    sh.nordvpn.set("notify", "off")
    sh.nordvpn.set("protection", "on")

    sh.nordvpn.connect()

    assert "Status: Connected" in sh.nordvpn.status(), "Status should show Connected"
    assert not settings.is_notify_enabled(), "Notifications should be disabled"
    assert settings.is_rtp_enabled(), "RTP should be enabled"

    assert settings.MSG_SET_DEFAULTS in sh.nordvpn.set.defaults(), "Defaults reset message should be shown"

    assert "Status: Disconnected" in sh.nordvpn.status(), "Status should show Disconnected after defaults reset"
    assert settings.app_has_defaults_settings(), "App should have default settings"
    assert "Account information" in sh.nordvpn.account(), "Account information should be displayed"


@pytest.mark.parametrize("nameserver", (dns.DNS_CASE_CUSTOM_SINGLE,))
@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_is_custom_dns_removed_after_setting_defaults_no_logout(vpn_protocol, nameserver):
    """Manual TC: LVPN-8748"""

    lib.set_vpn_protocol(vpn_protocol)

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


def test_settings_are_kept_after_reboot():
    # (set arguments, expected message, settings key, expected value after reboot)
    toggles = [
        (("firewall", "off"),         "Firewall has been successfully set to 'disabled'.",               "Firewall",               "disabled"),
        (("routing", "off"),          "Routing has been successfully set to 'disabled'.",                "Routing",                "disabled"),
        (("analytics", "off"),        "Analytics has been successfully set to 'disabled'.",              "User Consent",           "disabled"),
        (("protection", "on"),        "Real-time protection has been successfully set to 'enabled'.",    "Real-time protection",   "enabled"),
        (("notify", "off"),           "Notifications are set to 'disabled' successfully.",               "Notify",                 "disabled"),
        (("tray", "off"),             "Tray set to 'disabled' successfully.",                            "Tray",                   "disabled"),
        (("autoconnect", "on"),       "Auto-connect has been successfully set to 'enabled'.",            "Auto-connect",           "enabled"),
        (("lan-discovery", "on"),     "LAN Discovery has been successfully set to 'enabled'.",           "LAN Discovery",          "enabled"),
        (("arp-ignore", "off"),       "ARP ignore set to 'disabled' successfully.",                      "ARP Ignore",             "disabled"),
    ]

    for args, message, _, _ in toggles:
        assert message in sh.nordvpn.set(*args), f"Expected message not shown: {message}"

    assert "Firewall Mark has been successfully set to '0x1234'." in sh.nordvpn.set("fwmark", "0x1234"), "Failed to set firewall mark"

    daemon.restart()

    app_settings = settings.Settings()

    for _, _, key, expected in toggles:
        assert app_settings.get(key) == expected, f"{key} is incorrect after reboot '{expected}'"

    assert app_settings.get("Firewall Mark") == "0x1234", "Firewall mark is not kept after reboot"
    assert app_settings.get("DNS") == "disabled", "DNS must be disabled because RTP is enabled"

    # set DNS and reboot the system
    assert "DNS has been successfully set to '1.1.1.1'." in sh.nordvpn.set("dns", "1.1.1.1"), "Failed to set custom DNS"

    daemon.restart()

    app_settings = settings.Settings()
    assert app_settings.get("Real-time protection") == "disabled", "Real-time protection must be disabled, because of custom DNS"
    assert app_settings.get("DNS") == "1.1.1.1", "Custom DNS value is not kept after reboot"

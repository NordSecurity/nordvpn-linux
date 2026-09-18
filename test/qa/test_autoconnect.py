import random

import pytest
import sh

import lib
from lib import daemon, network, server, settings, IS_NIGHTLY
from lib.shell import sh_no_tty
from lib.dynamic_parametrize import dynamic_parametrize

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


def autoconnect_base_test(group):
    output = sh_no_tty.nordvpn.set.autoconnect.on(group)
    print(output)
    assert settings.MSG_AUTOCONNECT_ENABLE_SUCCESS in output, "Should show autoconnect enable success message"

    daemon.restart()
    daemon.wait_for_autoconnect()
    assert network.is_connected(), "Network should be connected after autoconnect"

    status_info = daemon.get_status_data()
    assert "Connected" in status_info["status"], "Status should show Connected"

    output = sh_no_tty.nordvpn.set.autoconnect.off()
    print(output)
    assert settings.MSG_AUTOCONNECT_DISABLE_SUCCESS in output, "Should show autoconnect disable success message"

    output = sh_no_tty.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output), "Disconnect should be successful"
    assert network.is_disconnected(), "Network should be disconnected"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_default(tech, proto, obfuscated):
    """Manual TC: LVPN-6779"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test("")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_not_autoconnect(tech, proto, obfuscated):
    """Manual TC: LVPN-6780"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    daemon.restart()
    assert network.is_disconnected(), "Network should be disconnected when autoconnect is off"


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "group",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.COUNTRIES + lib.COUNTRY_CODES],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{group}",
)
def test_autoconnect_to_country(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6781"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "group",
    ],
    ordered_source=[lib.TECHNOLOGIES],
    randomized_source=[lib.CITIES],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{group}",
)
def test_autoconnect_to_city(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6784"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "group",
    ],
    ordered_source=[lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER],
    randomized_source=[lib.ADDITIONAL_GROUPS],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{group}",
)
def test_autoconnect_to_additional_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6786"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@dynamic_parametrize(
    [
        "tech", "proto", "obfuscated", "group",
    ],
    ordered_source=[lib.NORDWHISPER_TECHNOLOGY],
    randomized_source=[lib.ADDITIONAL_GROUPS_NORDWHISPER],
    generate_all=IS_NIGHTLY,
    id_pattern="{tech}-{proto}-{obfuscated}-{group}",
)
def test_nordwhisper_autoconnect_to_additional_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6786"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.DEDICATED_IP_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_autoconnect_to_ovpn_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-563"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_autoconnect_to_obfuscated_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-410"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_virtual_country(tech, proto, obfuscated):
    """Manual TC: LVPN-8549"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set("virtual-location", "on")

    virtual_countries = lib.get_virtual_countries()
    assert len(virtual_countries) > 0, "Virtual countries should be available"
    country = random.choice(virtual_countries)

    autoconnect_base_test(country)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_virtual_country_disabled(tech, proto, obfuscated):
    """Manual TC: LVPN-8548"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    # fix in LVPN-8449
    # sh.nordvpn.set("virtual-location", "on")
    # virtual_countries = lib.get_virtual_countries()
    # assert len(virtual_countries) > 0
    # country = random.choice(virtual_countries)
    # until then chose a country that has only virtual server locations
    country = "AF"

    sh.nordvpn.set("virtual-location", "off")

    with pytest.raises(sh.ErrorReturnCode_1) as _:
        output = sh_no_tty.nordvpn.set.autoconnect.on(country)
        assert "Please enable virtual location access to connect to this server." in output, "Should show virtual location access error"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_unavailable_groups(tech, proto, obfuscated):
    """Manual TC: LVPN-8431"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh_no_tty.nordvpn.set.autoconnect.on(group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex), "Connection should be unsuccessful"

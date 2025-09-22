import random

import pytest
import sh

import lib
from lib import daemon, network, server, settings
from lib.shell import sh_no_tty


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


def autoconnect_base_test(group):
    output = sh_no_tty.nordvpn.set.autoconnect.on(group)
    print(output)
    assert settings.MSG_AUTOCONNECT_ENABLE_SUCCESS in output

    daemon.restart()
    daemon.wait_for_autoconnect()
    assert network.is_connected()

    output = sh_no_tty.nordvpn.set.autoconnect.off()
    print(output)
    assert settings.MSG_AUTOCONNECT_DISABLE_SUCCESS in output

    output = sh_no_tty.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()


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
    assert network.is_disconnected()


@pytest.mark.parametrize("group", lib.COUNTRIES + lib.COUNTRY_CODES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_country(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6781"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.CITIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_city(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6784"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_random_server_by_name(tech, proto, obfuscated):
    """Manual TC: LVPN-6782"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated)
    name = server_info.hostname.split(".")[0]

    autoconnect_base_test(name)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_standard_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-8424"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_autoconnect_to_additional_group(tech, proto, obfuscated, group):
    """Manual TC: LVPN-6786"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
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
    assert len(virtual_countries) > 0
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
        assert "Please enable virtual location access to connect to this server." in output


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_unavailable_groups(tech, proto, obfuscated):
    """Manual TC: LVPN-8431"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh_no_tty.nordvpn.set.autoconnect.on(group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_prevent_autoconnect_enable_to_non_obfuscated_servers_when_obfuscation_is_on(tech, proto, obfuscated):
    """Manual TC: LVPN-8581"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        server_name = server.get_hostname_by(group_name=group).hostname.split(".")[0]

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.set.autoconnect.on(server_name)
        print(ex.value)
        error_message = "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."
        assert error_message in str(ex.value)
        assert "Auto-connect: disabled" in sh.nordvpn.settings()
        daemon.restart()
        assert network.is_disconnected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_prevent_obfuscate_disable_with_autoconnect_enabled_to_obfuscated_server(tech, proto, obfuscated):
    """Manual TC: LVPN-5847"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    available_groups = str(sh.nordvpn.groups(_tty_out=False)).strip().split()

    for group in available_groups:
        server_name = server.get_hostname_by(tech, proto, obfuscated, group).hostname.split(".")[0]
        sh.nordvpn.set.autoconnect.on(server_name)

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
             sh.nordvpn.set.obfuscate.off()
        print(ex.value)
        error_message = "We couldn’t turn off obfuscation because your current auto-connect server is obfuscated by default. " \
            + "Set a different server for auto-connect, then turn off obfuscation."
        assert error_message in str(ex.value)
        assert "Obfuscate: enabled" in sh.nordvpn.settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_prevent_autoconnect_enable_to_obfuscated_servers_when_obfuscation_is_off(tech, proto, obfuscated):
    """Manual TC: LVPN-8591"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        server_name = server.get_hostname_by(group_name="Obfuscated_Servers").hostname.split(".")[0]
        sh.nordvpn.set.autoconnect.on(server_name)
    print(ex.value)
    error_message = "Turn on obfuscation to connect to obfuscated servers."
    assert error_message in str(ex.value)
    assert "Auto-connect: disabled" in sh.nordvpn.settings()

    daemon.restart()
    assert network.is_disconnected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_prevent_obfuscate_enable_with_autoconnect_set_to_nonobfuscated(tech, proto, obfuscated):
    """Manual TC: LVPN-5848"""

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    available_groups = str(sh.nordvpn.groups(_tty_out=False)).strip().split()

    for group in available_groups:
        if group == "Dedicated_IP":
            server_name = server.get_dedicated_ip().hostname.split(".")[0]
        else:
            server_name = server.get_hostname_by(tech, proto, obfuscated, group).hostname.split(".")[0]

        sh.nordvpn.set.autoconnect.on(server_name)

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
             sh.nordvpn.set.obfuscate.on()
        print(ex.value)
        error_message = "We couldn’t turn on obfuscation because the current auto-connect server doesn’t support it. " \
            + "Set a different server for auto-connect to use obfuscation."
        assert error_message in str(ex.value)
        assert "Obfuscate: disabled" in sh.nordvpn.settings()

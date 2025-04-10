import random

import pytest
import sh

import lib
from lib import daemon, info, logging, login, network, server, settings
from lib.shell import sh_no_tty


def setup_function(function):  # noqa: ARG001
    daemon.start()
    login.login_as("default")
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


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
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test("")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_not_autoconnect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    output = sh.nordvpn.set.autoconnect.off()
    print(output)

    daemon.restart()
    assert network.is_disconnected()


@pytest.mark.parametrize("group", lib.COUNTRIES + lib.COUNTRY_CODES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_country(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.CITIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_to_city(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_random_server_by_name(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated)
    name = server_info.hostname.split(".")[0]

    autoconnect_base_test(name)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_standard_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_to_additional_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.OVPN_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_to_ovpn_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_autoconnect_to_obfuscated_group(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    autoconnect_base_test(group)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_virtual_country(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set("virtual-location", "on")

    virtual_countries = lib.get_virtual_countries()
    assert len(virtual_countries) > 0
    country = random.choice(virtual_countries)

    autoconnect_base_test(country)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_autoconnect_virtual_country_disabled(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set("virtual-location", "on")

    virtual_countries = lib.get_virtual_countries()
    assert len(virtual_countries) > 0
    country = random.choice(virtual_countries)

    sh.nordvpn.set("virtual-location", "off")

    with pytest.raises(sh.ErrorReturnCode_1) as _:
        output = sh_no_tty.nordvpn.set.autoconnect.on(country)
        assert "Please enable virtual location access to connect to this server." in output


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_autoconnect_to_unavailable_groups(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh_no_tty.nordvpn.set.autoconnect.on(group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_prevent_autoconnect_enable_to_non_obfuscated_servers_when_obfuscation_is_on(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        server_name = server.get_hostname_by(group_id=group).hostname.split(".")[0]

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
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        server_name = server.get_hostname_by(group_id="Obfuscated_Servers").hostname.split(".")[0]
        sh.nordvpn.set.autoconnect.on(server_name)
    print(ex.value)
    error_message = "Turn on obfuscation to connect to obfuscated servers."
    assert error_message in str(ex.value)
    assert "Auto-connect: disabled" in sh.nordvpn.settings()

    daemon.restart()
    assert network.is_disconnected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OVPN_STANDARD_TECHNOLOGIES)
def test_prevent_obfuscate_enable_with_autoconnect_set_to_nonobfuscated(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    available_groups = str(sh.nordvpn.groups(_tty_out=False)).strip().split()

    for group in available_groups:
        if group == "Dedicated_IP":
            server_name = server.get_dedicated_ip()
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

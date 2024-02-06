import pytest
import sh
import timeout_decorator

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)


def setup_module(module):  # noqa: ARG001
    daemon.start()
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


autoconnect_on_parameters = [
    ("lt16", "on", "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."),
    ("uk2188", "off", "Turn on obfuscation to connect to obfuscated servers.")
]


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_obfuscate_nonobfucated(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.obfuscate("on")
        assert "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation." in str(ex.value)


@pytest.mark.skip(reason="LVPN-2119")
@timeout_decorator.timeout(40)
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


@timeout_decorator.timeout(40)
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

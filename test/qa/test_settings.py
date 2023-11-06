from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


autoconnectOnParameters = [
    ("lt16", "on", "Your selected server doesn’t support obfuscation. Choose a different server or turn off obfuscation."),
    ("uk2188", "off", "Turn on obfuscation to connect to obfuscated servers.")
]


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC1)
def test_obfuscate_nonobfucated(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.obfuscate("on")
        assert "Obfuscation is not available with the current technology. Change the technology to OpenVPN to use obfuscation." in str(ex.value)


@pytest.mark.skip(reason="LVPN-2119")
@timeout_decorator.timeout(40)
@pytest.mark.parametrize("server,obfuscated,errorMessage", autoconnectOnParameters)
def test_autoconnect_on_server_obfuscation_mismatch(server, obfuscated, errorMessage):
    lib.set_technology_and_protocol("openvpn", "tcp", obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.set.autoconnect.on(server)

    print(ex.value)
    assert errorMessage in str(ex.value)

    assert "Auto-connect: disabled" in sh.nordvpn.settings()

    daemon.restart()
    assert network.is_disconnected()

    sh.nordvpn.set.autoconnect.off()


setObfuscateParameters = [
    ("off", "lt16", "We couldn’t turn on obfuscation because the current auto-connect server doesn’t support it. Set a different server for auto-connect to use obfuscation."),
    ("on", "uk2188", "We couldn’t turn off obfuscation because your current auto-connect server is obfuscated by default. Set a different server for auto-connect, then turn off obfuscation.")
]


@timeout_decorator.timeout(40)
@pytest.mark.skip(reason="LVPN-2119")
@pytest.mark.parametrize("obfuscateInitialState,server,errorMessage", setObfuscateParameters)
def test_set_obfuscate_server_obfuscation_mismatch(obfuscateInitialState, server, errorMessage):
    lib.set_technology_and_protocol("openvpn", "tcp", obfuscateInitialState)

    output = sh.nordvpn.set.autoconnect.on(server)
    print(output)

    obfuscateExpectedState = "disabled"
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if obfuscateInitialState == "off":
            sh.nordvpn.set.obfuscate.on()
        else:
            obfuscateExpectedState = "enabled"
            sh.nordvpn.set.obfuscate.off()

    assert "Obfuscate: {}".format(obfuscateExpectedState) in sh.nordvpn.settings()

    assert errorMessage in str(ex.value)

    sh.nordvpn.set.autoconnect.off()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES_BASIC2 + lib.TECHNOLOGIES_BASIC1)
def test_set_technology(tech, proto, obfuscated):
    assert f"Technology is set to '{tech.upper()}' successfully." in sh.nordvpn.set.technology(tech)
    assert tech.upper() in sh.nordvpn.settings()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
def test_protocol_in_settings(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert proto.upper() in sh.nordvpn.settings()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_technology_set_options(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    
    ovpn_list = "obfuscate" in sh.nordvpn.set() and "protocol" in sh.nordvpn.set()

    if tech == "openvpn":
        assert ovpn_list
    else:
        assert not ovpn_list
import pytest
import sh

import lib
from lib import daemon, meshnet, settings, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


# This doesn't directly test meshnet, but it uses it
def test_allowlist_incoming_connection():
    my_ip = ssh_client.exec_command("echo $SSH_CLIENT").split()[0]

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_external_peer().hostname
    # Initiate ssh connection via mesh because we are going to lose the main connection
    ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
    ssh_client_mesh.connect()

    ssh_client_mesh.exec_command("nordvpn set killswitch on")
    # We should not have direct connection anymore after connecting to VPN
    with pytest.raises(sh.ErrorReturnCode_1):
        assert "icmp_seq=" not in sh.ping("-c", "1", "qa-peer")

        ssh_client_mesh.exec_command(f"nordvpn allowlist add subnet {my_ip}/32")
        # Direct connection should work again after allowlisting
        assert "icmp_seq=" in sh.ping("-c", "1", "qa-peer")
    ssh_client_mesh.exec_command("nordvpn set killswitch off")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES) # Only using standard technologies here because of "LVPN-4601 - Enabling Auto-connect disables Obfuscation"
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_in_2nd_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    
    sh.nordvpn.set.fwmark("0xe2f2")
    sh.nordvpn.set.killswitch("on")
    sh.nordvpn.set.tpl("on")
    sh.nordvpn.set.autoconnect("on")
    sh.nordvpn.set("lan-discovery", "on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in sh.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_tpl_enabled()
    assert settings.is_autoconnect_enabled()
    assert settings.is_lan_discovery_enabled()
    
    if tech == "openvpn":
        assert not settings.is_obfuscated_enabled()

    assert "Settings were successfully restored to defaults." in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
# This doesn't directly test meshnet, but it uses it
def test_set_defaults_when_logged_out_1st_set(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.fwmark("0xe2f2")
    sh.nordvpn.set.killswitch("on")
    sh.nordvpn.set("lan-discovery", "on")
    sh.nordvpn.set.analytics("off")
    sh.nordvpn.set.tpl("on")

    assert settings.is_meshnet_enabled()
    assert "0xe1f1" not in sh.nordvpn.settings()
    assert daemon.is_killswitch_on()
    assert settings.is_lan_discovery_enabled()
    assert not settings.are_analytics_enabled()
    assert settings.is_tpl_enabled()
    
    if obfuscated == "on":
        assert settings.is_obfuscated_enabled()
    else:
        assert not settings.is_obfuscated_enabled()

    sh.nordvpn.logout("--persist-token")

    assert "Settings were successfully restored to defaults." in sh.nordvpn.set.defaults()

    assert settings.app_has_defaults_settings()

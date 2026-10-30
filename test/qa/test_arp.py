import pytest

import lib
from lib import network, ssh
from lib.meshnet import TestUtils
from lib.shell import sh_no_tty

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module): # noqa: ARG001
    ssh_client.connect()
    TestUtils.allowlist_ssh(ssh_client, network.FWMARK)
    ssh_client.exec_command("sudo apt install iputils-arping -y")

def teardown_module(module): # noqa: ARG001
    ssh_client.disconnect()

@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS_BASIC1)
def test_arp_on_after_set_defaults_with_arp_ignore_disabled(vpn_protocol):
    lib.set_vpn_protocol(vpn_protocol)

    sh_no_tty.nordvpn.set("arp-ignore", "off")
    sh_no_tty.nordvpn.set.defaults()

    with lib.Defer(sh_no_tty.nordvpn.disconnect):
        sh_no_tty.nordvpn.connect()
        nordlynx_interface_ip = network.RouteInfo(sh_no_tty.ip.route.show.dev("nordlynx").stdout.decode()).src

        with pytest.raises(RuntimeError) as ex:
            ssh_client.exec_command(f"arping -c 1 -w 1 {nordlynx_interface_ip}")
        assert "Received 0 response" in str(ex.value), "ARP requests should fail"

        sh_no_tty.nordvpn.set("arp-ignore", "off")
        arping_output = ssh_client.exec_command(f"arping -c 1 -w 1 {nordlynx_interface_ip}")
        assert "Received 1 response" in arping_output, "ARP requests should succeed"

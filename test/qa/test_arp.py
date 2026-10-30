import pytest
import sh

import lib
from lib import network, ssh
from lib.meshnet import TestUtils

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module): # noqa: ARG001
    ssh_client.connect()
    TestUtils.allowlist_ssh(ssh_client, network.FWMARK)
    ssh_client.exec_command("sudo apt install iputils-arping -y")

def teardown_module(module): # noqa: ARG001
    ssh_client.disconnect()

@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS_BASIC2)
def test_arp_on_after_set_defaults_with_arp_ignore_disabled(vpn_protocol):
    lib.set_vpn_protocol(vpn_protocol)

    sh.nordvpn.set("arp-ignore", "off")
    sh.nordvpn.set.defaults()

    lib.set_vpn_protocol(vpn_protocol)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()
        nordtun_interface_ip = network.RouteInfo(sh.ip.route.show.dev("nordtun").stdout.decode()).src

        with pytest.raises(RuntimeError) as ex:
            ssh_client.exec_command(f"arping -c 1 -w 1 {nordtun_interface_ip}")
        assert "Received 0 response" in str(ex.value), "ARP requests should fail"

        sh.nordvpn.set("arp-ignore", "off")
        arping_output = ssh_client.exec_command(f"arping -c 1 -w 1 {nordtun_interface_ip}")
        assert "Received 1 response" in arping_output, "ARP requests should succeed"

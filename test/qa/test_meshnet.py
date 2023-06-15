from lib import daemon, info, logging, login, meshnet, ssh
import lib
import sh
import requests
import pytest
import timeout_decorator

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module):
    ssh_client.connect()
    daemon.install_peer(ssh_client)
    daemon.start()
    daemon.start_peer(ssh_client)
    login.login_as("default")
    login.login_as("qa-peer", ssh_client) # TODO: same account is used for everybody, tests can't be run in parallel
    lib.set_technology_and_protocol("nordlynx", "", "")
    sh.nordvpn.set.meshnet.on()
    ssh_client.exec_command("nordvpn set mesh on")
    # Ensure clean starting state
    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)


def teardown_module(module):
    ssh_client.exec_command("nordvpn set mesh off")
    sh.nordvpn.set.meshnet.off()
    ssh_client.exec_command("nordvpn logout --persist-token")
    sh.nordvpn.logout("--persist-token")
    daemon.stop_peer(ssh_client)
    daemon.stop()
    daemon.uninstall_peer(ssh_client)
    ssh_client.disconnect()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


def test_meshnet_connect():
    with lib.Defer(meshnet.remove_all_peers):
        meshnet.add_peer(ssh_client)
        # Ideally peer update should happen through Notification Center, but that doesn't work often
        sh.nordvpn.meshnet.peer.refresh()
        assert meshnet.is_peer_reachable(ssh_client)
    


def test_mesh_removed_machine_by_other():
    # find my token from cli
    mytoken = ""
    output =  sh.nordvpn.token()
    for ln in output.splitlines():
        if "Token:" in ln:
            _, mytoken = ln.split(None, 2)

    myname = meshnet.get_this_device(sh.nordvpn.mesh.peer.list())

    # find my machineid from api
    mymachineid = ""
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:'+mytoken,
    }
    response = requests.get('https://api.nordvpn.com/v1/meshnet/machines', headers=headers)
    for itm in response.json():
        if str(itm['hostname']) in myname:
            mymachineid = itm['identifier']

    # remove myself using api call
    headers = {
        'Accept': 'application/json',
        'Authorization': 'Bearer token:'+mytoken,
    }
    response = requests.delete('https://api.nordvpn.com/v1/meshnet/machines/'+mymachineid, headers=headers)

    # machine not found error should be handled by disabling meshnet
    try:
        output = sh.nordvpn.mesh.peer.list()
    except Exception as e:
        assert "Meshnet is not enabled." in str(e)

    sh.nordvpn.set.meshnet.on() # enable back on for other tests


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
# This doesn't directly test meshnet, but it uses it
def test_whitelist_incoming_connection():
    with lib.Defer(meshnet.remove_all_peers):
        meshnet.add_peer(ssh_client)
        # Ideally peer update should happen through Notification Center, but that doesn't work often
        sh.nordvpn.meshnet.peer.refresh()
        my_ip = ssh_client.exec_command("echo $SSH_CLIENT").split()[0]

        peer_hostname = meshnet.get_this_device(ssh_client.exec_command("nordvpn mesh peer list"))
        # Initiate ssh connection via mesh because we are going to lose the main connection
        ssh_client_mesh = ssh.Ssh(peer_hostname, "root", "root")
        ssh_client_mesh.connect()
        with lib.Defer(ssh_client_mesh.disconnect):
            ssh_client_mesh.exec_command("nordvpn c")
            with lib.Defer(lambda: ssh_client_mesh.exec_command("nordvpn d")):
                # We should not have direct connection anymore after connecting to VPN
                with pytest.raises(sh.ErrorReturnCode_1) as ex:
                    assert "icmp_seq=" not in sh.ping("-c", "1", "qa-peer")

                    ssh_client_mesh.exec_command(f"nordvpn whitelist add subnet {my_ip}/32")
                    with lib.Defer(lambda: ssh_client_mesh.exec_command("nordvpn whitelist remove all")):
                        # Direct connection should work again after whitelisting
                        assert "icmp_seq=" in sh.ping("-c", "1", "qa-peer")


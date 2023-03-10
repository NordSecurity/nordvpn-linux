from lib import daemon, info, logging, login, meshnet, ssh
import lib
import sh
import requests

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
    meshnet.add_peer(ssh_client)
    # Ideally peer update should happen through Notification Center, but that doesn't work often
    sh.nordvpn.meshnet.peer.refresh()
    assert meshnet.is_peer_reachable(ssh_client)
    meshnet.remove_all_peers()


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

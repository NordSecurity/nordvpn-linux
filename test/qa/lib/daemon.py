"""Functions to make it easier to interact with nordvpnd."""
import glob
import os
import socket
import time

import sh

from . import logging, ssh
from lib.shell import sh_no_tty


def _rewrite_log_path():
    project_root = os.environ["WORKDIR"].replace("/", "\\/")
    pattern = f"s/^LOGFILE=.*/LOGFILE={project_root}\\/dist\\/logs\\/daemon.log/"
    # this fn is executed only in docker (below line would not work under snap)
    sh.sudo.sed("-i", pattern, "/etc/init.d/nordvpn")


# returns True on SystemD distros
def is_init_systemd():
    return "systemd" in sh.ps("--no-headers", "-o", "comm", "1")


def is_under_snap():
    return "snap" in sh.which("nordvpn")


def is_connected() -> bool:
    """Returns True when connected to VPN server."""
    try:
        return "Connected" in sh.nordvpn.status()
    except sh.ErrorReturnCode:
        return False


def is_disconnected() -> bool:
    """Returns True when not connected to VPN."""
    try:
        print(sh_no_tty.nordvpn.status())
        return "Disconnected" in sh.nordvpn.status()
    except sh.ErrorReturnCode as ex:
        logging.log(data=f"is_disconnected: {ex}")
        return True


def is_killswitch_on():
    """Return True when Killswitch is activated."""
    try:
        return "Kill Switch: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


def install_peer(ssh_client: ssh.Ssh):
    """Installs nordvpn in peer."""
    project_root = os.environ["WORKDIR"]
    deb_path = glob.glob(f'{project_root}/dist/app/deb/*amd64.deb')[0]
    ssh_client.send_file(deb_path, '/tmp/nordvpn.deb')
    # TODO: Install required dependencies during qa-peer image build, then replace with 'dpkg -i /tmp/nordvpn.deb'
    ssh_client.exec_command('sudo apt-get update')
    ssh_client.exec_command('sudo apt install -y /tmp/nordvpn.deb')


def uninstall_peer(ssh_client: ssh.Ssh):
    """Uninstalls nordvpn in peer."""
    # TODO: Replace with 'dpkg --purge nordvpn'
    ssh_client.exec_command('sudo apt remove -y nordvpn')


def start():
    """Starts daemon and blocks until it is actually started."""
    if is_under_snap():
        #sh.sudo.snap("start", "nordvpn")
        os.popen("sudo snap start nordvpn").read()
    elif is_init_systemd():
        #sh.sudo.systemctl.start.nordvpnd()
        os.popen("sudo systemctl start nordvpn").read()
    else:
        # call to init.d returns before the daemon is actually started
        _rewrite_log_path()
        sh.sudo("/etc/init.d/nordvpn", "start")
    while not is_running():
        time.sleep(1)


def start_peer(ssh_client: ssh.Ssh):
    """Starts daemon in peer and blocks until it is actually started."""
    ssh_client.exec_command("sudo /etc/init.d/nordvpn start")
    time.sleep(1)
    while not is_peer_running(ssh_client):
        time.sleep(1)


def stop():
    """Stops the daemon and blocks until it is actually stopped."""
    if is_under_snap():
        #sh.sudo.snap("stop", "nordvpn")
        os.popen("sudo snap stop nordvpn").read()
    elif is_init_systemd():
        #sh.sudo.systemctl.stop.nordvpnd()
        os.popen("sudo systemctl stop nordvpn").read()
    else:
        # call to init.d returns before the daemon is actually stopped
        sh.sudo("/etc/init.d/nordvpn", "stop")
        while is_running():
            time.sleep(1)


def stop_peer(ssh_client: ssh.Ssh):
    """Stops the daemon in peer and blocks until it is actually stopped."""
    ssh_client.exec_command("sudo /etc/init.d/nordvpn stop")
    while is_peer_running(ssh_client):
        time.sleep(1)


# restarts the daemon and blocks until it is actually restarted
def restart():
    stop()
    if is_init_systemd():
        # The default limit is to allow 5 restarts in a 10sec period.
        # There are 100ms intervals between restarts.
        time.sleep(2)
    start()


# retrieving links inside this function creates a race condition,
# therefore it is safer to provide them as arguments
def wait_for_reconnect(links: list[tuple[int, str]]):
    logging.log("waiting for reconnect")
    while True:
        got = socket.if_nameindex()
        if len(got) != len(links):  # old tunnel is gone
            continue
        if got != links:  # new tunnel appeared
            logging.log(got)
            if list(filter(lambda x: "nordvpn-wg" in x, got)):
                # not yet connected to actual VPN, this is just a test interface
                continue
            time.sleep(2)
            break


def wait_for_autoconnect():
    while not is_connected():
        time.sleep(1)


# returns True when daemon is running
def is_running():
    try:
        sh.nordvpn.status()
    except sh.ErrorReturnCode_1:
        return False
    else:
        return True


# returns True when daemon is running in peer
def is_peer_running(ssh_client: ssh.Ssh) -> bool:
    try:
        ssh_client.exec_command("nordvpn status")
        return True
    except Exception:  # noqa: BLE001
        return False


def get_unavailable_groups():
    """Returns groups that are not available with current connection settings."""
    all_groups = ['Africa_The_Middle_East_And_India',
                  'Asia_Pacific',
                  'Dedicated_IP',
                  'Double_VPN',
                  'Europe',
                  'Obfuscated_Servers',
                  'Onion_Over_VPN',
                  'P2P',
                  'Standard_VPN_Servers',
                  'The_Americas']

    current_groups = str(sh.nordvpn.groups(_tty_out=False)).strip().split()

    return set(all_groups) - set(current_groups)


def get_status_data() -> dict:
    lines = sh.nordvpn.status(_tty_out=False).strip().split('\n')
    colon_separated_pairs = (element.split(':') for element in lines)
    formatted_pairs = {(key.lower(), value.strip()) for key, value in colon_separated_pairs}
    return dict(formatted_pairs)

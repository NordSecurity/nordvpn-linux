"""Functions to make it easier to interact with nordvpnd."""
from lib import ssh
import logging
import os
import sh
import socket
import time
import glob
from typing import List, Tuple


def _rewrite_log_path():
    project_root = os.environ["CI_PROJECT_DIR"].replace("/", "\/")
    pattern = f"s/^LOGFILE=.*/LOGFILE={project_root}\/dist\/logs\/daemon.log/"
    sh.sudo.sed("-i", pattern, "/etc/init.d/nordvpn")


# returns True on SystemD distros
def is_init_systemd():
    return "systemd" in sh.ps("--no-headers", "-o", "comm", "1")


def is_connected() -> bool:
    """returns True when connected to VPN server"""
    try:
        return "Connected" in sh.nordvpn.status()
    except sh.ErrorReturnCode:
        return False


def is_disconnected() -> bool:
    """returns True when not connected to VPN"""
    try:
        return "Disconnected" in sh.nordvpn.status()
    except sh.ErrorReturnCode:
        return False


# return True when Killswitch is activated
def is_killswitch_on():
    try:
        return "Kill Switch: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


# return True when IPv6 is activated
def is_ipv6_on():
    try:
        return "IPv6: enabled" in sh.nordvpn.settings()
    except sh.ErrorReturnCode:
        return False


def install_peer(ssh_client: ssh.Ssh):
    """installs nordvpn in peer"""
    project_root = os.environ["CI_PROJECT_DIR"]
    deb_path = glob.glob(f'{project_root}/dist/app/deb/*amd64.deb')[0]
    ssh_client.send_file(deb_path, '/tmp/nordvpn.deb')
    ssh_client.exec_command('apt install -y /tmp/nordvpn.deb')


def uninstall_peer(ssh_client: ssh.Ssh):
    """uninstalls nordvpn in peer"""
    ssh_client.exec_command('apt remove -y nordvpn')


def start():
    """starts daemon and blocks until it is actually started"""
    if is_init_systemd():
        return sh.sudo.systemctl.start.nordvpnd()
    # call to init.d returns before the daemon is actually started
    _rewrite_log_path()
    sh.sudo("/etc/init.d/nordvpn", "start")
    while not is_running():
        time.sleep(1)


def start_peer(ssh_client: ssh.Ssh):
    """starts daemon in peer and blocks until it is actually started"""
    ssh_client.exec_command("/etc/init.d/nordvpn start")
    while not is_peer_running(ssh_client):
        time.sleep(1)


def stop():
    """stops the daemon and blocks until it is actually stopped"""
    if is_init_systemd():
        return sh.sudo.systemctl.stop.nordvpnd()
    # call to init.d returns before the daemon is actually stopped
    sh.sudo("/etc/init.d/nordvpn", "stop")
    while is_running():
        time.sleep(1)


def stop_peer(ssh_client: ssh.Ssh):
    """stops the daemon in peer and blocks until it is actually stopped"""
    ssh_client.exec_command("/etc/init.d/nordvpn stop")
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
def wait_for_reconnect(links: List[Tuple[int, str]]):
    logging.info("waiting for reconnect")
    while True:
        got = socket.if_nameindex()
        if len(got) != len(links):  # old tunnel is gone
            continue
        if got != links:  # new tunnel appeared
            logging.debug(got)
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
    except:
        return False
    else:
        return True

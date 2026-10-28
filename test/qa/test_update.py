import glob
import os
import re
import shutil

import pytest
import requests
import sh

import lib
from lib import daemon, fileshare, login, meshnet, network, poll, ssh, logging
from lib.shell import sh_no_tty
from test_connect import connect_base_test, disconnect_base_test

PROJECT_ROOT = os.environ['WORKDIR']
DEB_PATH = glob.glob(f'{PROJECT_ROOT}/dist/app/deb/nordvpn_*amd64.deb')[0]
PROD_REPO_PACKAGES_URL = "https://repo.nordvpn.com/deb/nordvpn/debian/dists/stable/main/binary-amd64/Packages"

def _extract_release_version(version: str) -> tuple[int, int, int] | None:
    match = re.match(r"(\d+)\.(\d+)\.(\d+)", version)
    return tuple(int(part) for part in match.groups()) if match else None


def _published_app_versions(packages_index: str) -> set[tuple[int, int, int]]:
    versions = set()
    for stanza in packages_index.split("\n\n"):
        fields = dict(re.findall(r"^(Package|Version): (.+)$", stanza, re.MULTILINE))
        if fields.get("Package") == "nordvpn" and (version := _extract_release_version(fields.get("Version", ""))):
            versions.add(version)
    return versions


def _previous_app_version() -> str:
    """Return the newest app version published in the prod repo that is older than the version under test."""

    current = _extract_release_version(sh.dpkg_deb("-f", DEB_PATH, "Version").strip())
    assert current, f"cannot parse version of {DEB_PATH}"

    response = requests.get(PROD_REPO_PACKAGES_URL, timeout=30)
    response.raise_for_status()

    older = [version for version in _published_app_versions(response.text) if version < current]
    assert older, f"no nordvpn version older than {current} is published in {PROD_REPO_PACKAGES_URL}"
    return ".".join(str(num) for num in max(older))

class TestData:
    INVOLVES_MESHNET = None

class ProductionApplicationData:
    APP_VERSION = None

ssh_client = ssh.Ssh("qa-peer", "root", "root")

def setup_module(module):  # noqa: ARG001
    os.system("sudo mkdir -p -m 0777 /home/qa/Downloads")

    ssh_client.connect()
    ssh_client.exec_command("mkdir -p /root/Downloads")


def setup_function(function):  # noqa: ARG001
    TestData.INVOLVES_MESHNET = any(keyword in os.environ["PYTEST_CURRENT_TEST"] for keyword in ["meshnet", "fileshare"])

    sh.sudo.apt.purge("-y", "nordvpn")

    # in order to simulate actual upgrade, always install current-but-one version
    # the exact version to be installed, is the newest one older than the tested build, which is still published in the prod repo.
    # finally, use the fetched previous version and provide it to the installation script
    previous_version = _previous_app_version()
    logging.log(data=f"Installing previous production version: {previous_version}")
    sh.sh("-s", "--", "-n", "-v", previous_version, _in=sh.curl("-sSf", "https://downloads.nordcdn.com/apps/linux/install.sh"))
    ProductionApplicationData.APP_VERSION = sh.nordvpn("-v").split()[2]

    daemon.start()

    daemon.stop() # TODO: LVPN-6403
    sh.sudo.apt.install(DEB_PATH, "-y")

    daemon.start() # TODO: LVPN-6403

    # login into the app after update, because if user didn't agree with the consent it will be logged out at update
    login.login_as("default")

    if TestData.INVOLVES_MESHNET:
        sh.nordvpn.set.notify.off()
        assert "Meshnet is set to 'enabled' successfully." in sh.nordvpn.set.meshnet.on(), "Meshnet should be successfully enabled"

        meshnet.remove_all_peers()

        daemon.install_peer(ssh_client)
        daemon.start_peer(ssh_client)
        login.login_as("default", ssh_client)
        ssh_client.exec_command("nordvpn set notify off")
        ssh_client.exec_command("nordvpn set mesh on")
        sh.nordvpn.mesh.peer.refresh()

        meshnet.wait_for_peers_connection(ssh_client)


def teardown_function(function):  # noqa: ARG001
    if TestData.INVOLVES_MESHNET:
        logging.log(data="Disable meshnet")
        ssh_client.exec_command("rm -rf /root/Downloads/*")

        sh.nordvpn.set.mesh.off()
        ssh_client.exec_command("nordvpn set mesh off")

        daemon.stop_peer(ssh_client)

        dest_logs_path = f"{PROJECT_ROOT}/dist/logs"
        meshnet.download_remote_peer_logs(ssh_client=ssh_client, dest_logs_path=dest_logs_path)
        shutil.copy("/home/qa/.cache/nordvpn/norduserd.log", dest_logs_path)
        shutil.copy("/home/qa/.cache/nordvpn/nordfileshare.log", dest_logs_path)

        daemon.uninstall_peer(ssh_client)
    daemon.stop() # TODO: LVPN-6403
    assert network.is_disconnected(), "Network should be disconnected after daemon stop"


def test_meshnet_available_after_update():
    """Manual TC: LVPN-3204"""

    meshnet_help_page = sh_no_tty.nordvpn.meshnet("--help")
    assert "Learn more: https://meshnet.nordvpn.com/" in meshnet_help_page, "Meshnet help page should contain meshnet.nordvpn.com reference"

    parsed_peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())

    local_hostname = parsed_peer_list.get_this_device().hostname
    ssh_client.exec_command(f"nordvpn mesh peer routing allow {local_hostname}")
    sh.nordvpn.mesh.peer.refresh()

    peer_hostname = parsed_peer_list.get_internal_peer().hostname
    output = sh.nordvpn.mesh.peer.connect(peer_hostname)
    assert meshnet.is_connect_successful(output, peer_hostname), "Meshnet peer connect should be successful"
    assert daemon.is_connected(), "Daemon should be connected after meshnet peer connect"
    assert network.is_available(), "Network should be available when connected to meshnet peer"

    output = sh.nordvpn.disconnect()
    assert lib.is_disconnect_successful(output), "Disconnect should be successful"
    assert not daemon.is_connected(), "Daemon should be disconnected after disconnect"
    assert network.is_available(), "Network should be available after disconnect"


@pytest.mark.xfail(condition=meshnet.is_meshnet_test_disabled_from_run(), reason="Run only in nightly")
def test_fileshare_available_after_update():
    """Manual TC: LVPN-3205"""

    fileshare_help_page = sh.nordvpn.fileshare("--help", _tty_out=False)
    assert "Learn more: https://meshnet.nordvpn.com/features/sharing-files-in-meshnet" in fileshare_help_page, "Fileshare help page should contain meshnet sharing reference"

    wdir = fileshare.create_directory(5)

    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().hostname

    fileshare.start_transfer(peer_hostname, wdir.dir_path)

    for remote_transfer_id, error_message in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=5):  # noqa: B007
        if remote_transfer_id is not None:
            break

    assert remote_transfer_id is not None, error_message

    ssh_client.exec_command(f"nordvpn fileshare accept {remote_transfer_id}")

    fileshare.files_from_transfer_exist_in_filesystem(remote_transfer_id, [wdir], ssh_client)


@pytest.mark.parametrize("vpn_protocol", lib.VPN_PROTOCOLS)
def test_quick_connect_after_update(vpn_protocol):
    """Manual TC: LVPN-8506"""

    lib.set_vpn_protocol(vpn_protocol)

    connect_base_test(vpn_protocol)
    disconnect_base_test()


def test_changelog_after_update():
    """Manual TC: LVPN-4152"""

    if ProductionApplicationData.APP_VERSION in sh.nordvpn("-v"):
        pytest.skip("Changelog not implemented yet.")

    changelog_path = "/usr/share/doc/nordvpn/changelog.Debian.gz"
    changelog = sh.dpkg_parsechangelog("-l", changelog_path)

    # take version without checksum as it will never be present in the changelog
    nordvpn_version = sh.nordvpn("-v").split()[2].split('+')[0]
    assert nordvpn_version in changelog, f"Version {nordvpn_version} should be present in changelog"
    assert "*" in changelog, "Changelog should contain entry markers"

import contextlib
import os
import random
import re
import shutil
import subprocess
import tempfile
import threading
import time

import psutil
import pytest
import sh

from lib import daemon, fileshare, info, logging, login, meshnet, poll, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")

# for snap testing, make this path from current folder e.g. ./tmp/testfiles
workdir = "/tmp/testfiles"
test_files = ["testing_fileshare_0.txt", "testing_fileshare_1.txt", "testing_fileshare_2.txt", "testing_fileshare_3.txt"]

default_download_directory = "/home/qa/Downloads"

def setup_module(module):  # noqa: ARG001
    os.makedirs("/home/qa/.config/nordvpn", exist_ok=True)
    os.makedirs("/home/qa/.cache/nordvpn", exist_ok=True)
    daemon.start()
    login.login_as("default")

    # temporary hack for autoaccept tests, we create a default download directory
    # will be remove once default download directory setting is implemented
    os.system(f"sudo mkdir -p -m 0777 {default_download_directory}")

    sh.nordvpn.set.notify.off()
    sh.nordvpn.set.meshnet.on()
    # Ensure clean starting state
    meshnet.remove_all_peers()

    ssh_client.connect()
    daemon.install_peer(ssh_client)
    daemon.start_peer(ssh_client)
    login.login_as("default", ssh_client)

    ssh_client.exec_command(f"mkdir -p {workdir}")
    ssh_client.exec_command(f"chmod 0777 {workdir}")
    ssh_client.exec_command("nordvpn set notify off")
    ssh_client.exec_command("nordvpn set mesh on")

    ssh_client.exec_command("nordvpn mesh peer refresh")
    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer()
    assert meshnet.is_peer_reachable(ssh_client, peer)

    if not os.path.exists(workdir):
        os.makedirs(workdir)

    message = "testing fileshare"
    for file in test_files:
        filepath = f"{workdir}/{file}"
        with open(filepath, "w") as f:
            f.write(message)

    ssh_client.exec_command("mkdir -p /root/Downloads")


def teardown_module(module):  # noqa: ARG001
    dest_logs_path = f"{os.environ['WORKDIR']}/dist/logs"
    # Preserve other peer log

    ssh_client.download_file("/var/log/nordvpn/daemon.log", f"{dest_logs_path}/other-peer-daemon.log")

    shutil.copy("/home/qa/.cache/nordvpn/norduserd.log", dest_logs_path)
    shutil.copy("/home/qa/.cache/nordvpn/nordfileshare.log", dest_logs_path)
    ssh_client.exec_command("nordvpn set mesh off")
    ssh_client.exec_command("nordvpn set notify on")
    ssh_client.exec_command("nordvpn logout --persist-token")
    daemon.stop_peer(ssh_client)
    daemon.uninstall_peer(ssh_client)
    ssh_client.disconnect()

    sh.nordvpn.set.meshnet.off()
    sh.nordvpn.set.notify.on()
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()
    fileshare.cancel_not_finished_transfers()
    ssh_client.exec_command(f"rm -rf {workdir}/*")


@pytest.mark.parametrize("accept_directories",
                         [["nested", "outer"],
                          ["nested"],
                          ["outer", "nested/inner"],
                          ["nested/inner"]])
def test_accept(accept_directories):
    address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().ip

    # Check peer list on both ends
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    logging.log(data="------------------11----------------------------------")
    logging.log(data=output)
    logging.log(data="------------------------------------------------------")
    output = ssh_client.exec_command("nordvpn mesh peer list")
    logging.log(data="------------------11r---------------------------------")
    logging.log(data=output)
    logging.log(data="------------------------------------------------------")

    # .
    # ├── nested
    # │   ├── file
    # │   └── inner
    # │       └── file
    # └── outer
    #     └── file

    nested_dir = "nested"
    inner_dir = "inner"
    outer_dir = "outer"
    filename = "file"

    ssh_client.exec_command(f"mkdir -p {workdir}/{nested_dir}")
    ssh_client.exec_command(f"echo > {workdir}/{nested_dir}/{filename}")
    ssh_client.exec_command(f"mkdir -p {workdir}/{nested_dir}/{inner_dir}")
    ssh_client.exec_command(f"echo > {workdir}/{nested_dir}/{inner_dir}/{filename}")
    ssh_client.exec_command(f"mkdir -p {workdir}/{outer_dir}")
    ssh_client.exec_command(f"echo > {workdir}/{outer_dir}/{filename}")

    transfer_files = [f"{nested_dir}/{filename}", f"{nested_dir}/{inner_dir}/{filename}", f"{outer_dir}/{filename}"]

    # Check peer list on both ends
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    logging.log(data="------------------22----------------------------------")
    logging.log(data=output)
    logging.log(data="------------------------------------------------------")
    output = ssh_client.exec_command("nordvpn mesh peer list")
    logging.log(data="------------------22r---------------------------------")
    logging.log(data=output)
    logging.log(data="------------------------------------------------------")

    # accept entire transfer
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {workdir}/{nested_dir} {workdir}/{outer_dir}")

    local_transfer_id = None
    error_message = None
    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):  # noqa: B007
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    peer_transfer_id = fileshare.get_last_transfer(ssh_client=ssh_client)

    sh.nordvpn.fileshare.accept("--background", "--path", "/tmp", local_transfer_id, *accept_directories).stdout.decode("utf-8")

    def predicate(file_entry: str) -> bool:
        file_entry_columns = file_entry.split(' ')
        for directory in accept_directories:
            if file_entry_columns[0].startswith(directory) and ("downloaded" in file_entry or "uploaded" in file_entry):
                return True

        return "canceled" in file_entry

    def check_files_status_receiver():
        tsfr = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
        return fileshare.for_all_files_in_transfer(tsfr, transfer_files, predicate), tsfr

    receiver_files_status_ok = False
    transfer = None
    for receiver_files_status_ok, transfer in poll(check_files_status_receiver, attempts=10):  # noqa: B007
        if receiver_files_status_ok is True:
            break

    assert receiver_files_status_ok is True, f"invalid file status on receiver side, transfer {transfer}, files {accept_directories} should be downloaded"

    def check_files_status_sender():
        tsfr = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
        return fileshare.for_all_files_in_transfer(tsfr, transfer_files, predicate), tsfr

    sender_files_status_ok = False
    for sender_files_status_ok, transfer in poll(check_files_status_sender, attempts=10):  # noqa: B007
        if sender_files_status_ok is True:
            break

    assert sender_files_status_ok is True, f"invalid file status on sender side, transfer {transfer}, files {accept_directories} should be uploaded"


@pytest.mark.parametrize("path_flag", [True, False], ids=["accept_custom_path", "accept_downloads"])
@pytest.mark.parametrize("background_accept", ["", "--background"], ids=["accept_int", "accept_bg"])
@pytest.mark.parametrize("background_send", [True, False], ids=["send_bg", "send_int"])
@pytest.mark.parametrize("filesystem_entity", list(fileshare.FileSystemEntity), ids = [f"send_{entity.value}" for entity in list(fileshare.FileSystemEntity)])
def test_fileshare_transfer(filesystem_entity: fileshare.FileSystemEntity, background_send: bool, path_flag: str, background_accept: str):
    peer_name = random.choice(list(meshnet.PeerName)[:-1])
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().get_peer_name(peer_name)

    # .
    # ├── wdir
    # │   └── wfolder
    # │       ├── file
    # │       └── file

    wdir = fileshare.create_directory(0)
    wfolder = fileshare.create_directory(2, "tmp", parent_dir=wdir.dir_path, file_size=128)

    if filesystem_entity == fileshare.FileSystemEntity.FILE:
        filepath = wfolder.paths[0]
    elif filesystem_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        filepath = wfolder.dir_path
    elif filesystem_entity == fileshare.FileSystemEntity.DIRECTORY_WITH_FOLDERS:
        filepath = wdir.dir_path
    elif filesystem_entity == fileshare.FileSystemEntity.FILES:
        filepath = wfolder.paths

    if background_send:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, filepath)
        output = command_handle.stdout.decode("utf-8")
        assert len(re.findall(fileshare.SEND_NOWAIT_SUCCESS_MSG_PATTERN, output)) > 0
    elif filesystem_entity == fileshare.FileSystemEntity.FILES:
        command_handle = fileshare.start_transfer(peer_address, *filepath)
    else:
        command_handle = fileshare.start_transfer(peer_address, filepath)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = None

    for last_peer_transfer_id, _ in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if last_peer_transfer_id is not None:
            peer_transfer_id = last_peer_transfer_id
            break

    assert peer_transfer_id is not None, "transfer was not received by peer"

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "request sent" in fileshare.find_transfer_by_id(transfers_local, local_transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "waiting for download" in fileshare.find_transfer_by_id(transfers_remote, peer_transfer_id)

    transfer_progress_local = fileshare.TransferProgressValidationThread(local_transfer_id, fileshare.TransferState.UPLOADING, None)
    transfer_progress_local.start()

    transfer_progress_remote = fileshare.TransferProgressValidationThread(local_transfer_id, fileshare.TransferState.DOWNLOADING, ssh_client)
    transfer_progress_remote.start()

    if path_flag:
        peer_filepath = "/tmp/"
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept {background_accept} --path {peer_filepath} {peer_transfer_id}")
    else:
        peer_filepath = "~/Downloads/"
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept {background_accept} {peer_transfer_id}")

    transfer_progress_local.join()
    transfer_progress_remote.join()

    assert transfer_progress_local.transfer_progress_valid
    assert transfer_progress_remote.transfer_progress_valid

    time.sleep(1)

    if not background_send:
        assert fileshare.validate_transfer_progress(command_handle.stdout.decode())
        assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, command_handle.stdout.decode())) == 1

    if not background_accept:
        assert fileshare.validate_transfer_progress(t_progress_interactive)
        assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, t_progress_interactive)) == 1

    assert fileshare.files_from_transfer_exist_in_filesystem(local_transfer_id, [wfolder], ssh_client)

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "completed" in fileshare.find_transfer_by_id(transfers_local, local_transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers_remote, peer_transfer_id)

    assert "uploaded" in sh.nordvpn.fileshare.list(local_transfer_id)
    shutil.rmtree(wdir.dir_path)
    ssh_client.exec_command(f"sudo rm -rf {peer_filepath}/*tmp*")


@pytest.mark.parametrize("path_flag", [True, False], ids=["accept_custom_path", "accept_downloads"])
@pytest.mark.parametrize("background_accept", ["", "--background"], ids=["accept_int", "accept_bg"])
@pytest.mark.parametrize("background_send", [True, False], ids=["send_bg", "send_int"])
def test_fileshare_transfer_multiple_files(background_send: bool, path_flag: str, background_accept: str):
    peer_name = random.choice(list(meshnet.PeerName)[:-1])
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().get_peer_name(peer_name)

    file_size = 22
    dir1 = fileshare.create_directory(0, "1")
    dir2 = fileshare.create_directory(2, "2", dir1.dir_path)
    dir3 = fileshare.create_directory(2, "3", file_size=file_size)
    dir4 = fileshare.create_directory(2, "4" ,file_size=file_size)

    # .
    # ├── dir1 - send this
    # │   └── dir2
    # │       └── file_1
    # │       └── file_2
    # │── dir3 - send this
    # │    └── file_1
    # │    └── file_2
    # │── dir4
    # │    └── file_1 - send this
    # │    └── file_2 - send this

    # transfer dir4 as individual files, i.e /<dir4>/<file1> /<dir4>/<file2>
    files_to_transfer = [dir1.dir_path, dir3.dir_path, *dir4.paths]

    if background_send:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, *files_to_transfer)
        output = command_handle.stdout.decode("utf-8")
        assert len(re.findall(fileshare.SEND_NOWAIT_SUCCESS_MSG_PATTERN, output)) > 0
    else:
        command_handle = fileshare.start_transfer(peer_address, *files_to_transfer)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()

    for last_peer_transfer_id, _ in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if last_peer_transfer_id is not None:
            peer_transfer_id = last_peer_transfer_id
            break

    assert peer_transfer_id is not None, "transfer was not received by peer"

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "request sent" in fileshare.find_transfer_by_id(transfers_local, local_transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "waiting for download" in fileshare.find_transfer_by_id(transfers_remote, peer_transfer_id)

    files_in_transfer = dir1.transfer_paths + dir2.transfer_paths + dir3.filenames

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "request sent" in file_entry)

    if path_flag:
        peer_filepath = workdir
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept {background_accept} --path {peer_filepath} {peer_transfer_id}")
    else:
        peer_filepath = "~/Downloads/"
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept {background_accept} {peer_transfer_id}")

    for transfers_done in poll(
        lambda: (
            "completed" in fileshare.get_transfer(local_transfer_id) and
            "completed" in fileshare.get_transfer(local_transfer_id, ssh_client)
        )
    ):
        if transfers_done:
            break

    time.sleep(1)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "uploaded" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "downloaded" in file_entry)

    if not background_send:
        assert fileshare.validate_transfer_progress(command_handle.stdout.decode())
        assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, command_handle.stdout.decode())) == 1

    if not background_accept:
        assert fileshare.validate_transfer_progress(t_progress_interactive)
        assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, t_progress_interactive)) == 1

    assert fileshare.files_from_transfer_exist_in_filesystem(local_transfer_id, [dir2, dir3, dir4], ssh_client)

    for entity in [dir1, dir3, dir4]:
        shutil.rmtree(entity.dir_path)

    ssh_client.exec_command(f"sudo rm -rf {peer_filepath}/*tmp*")

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id)


@pytest.mark.parametrize("accept_entity", list(fileshare.FileSystemEntity)[:-1], ids = [f"accept_{entity.value}" for entity in list(fileshare.FileSystemEntity)[:-1]])
@pytest.mark.parametrize("background", [True, False], ids=["send_bg", "send_int"])
def test_fileshare_transfer_multiple_files_selective_accept(background: bool, accept_entity: fileshare.FileSystemEntity):
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    # .
    # ├── wfolder_1
    # │   ├── file
    # │   └── file
    # ├── wfolder_2
    # │   ├── file
    # │   └── file
    # ├── wfolder_3
    # │   ├── file
    # │   └── file
    # ├── wdir_1
    # │   └── wfolder_4
    # │       └── file
    # │       └── file
    # ├── wdir_2
    # │   └── wfolder_5
    # │       └── file
    # │       └── file

    file_size = 128

    wfolder_1 = fileshare.create_directory(2, "1", file_size=file_size)
    wfolder_2 = fileshare.create_directory(2, "2", file_size=file_size)
    wfolder_3 = fileshare.create_directory(2, "3")

    wdir_1 = fileshare.create_directory(0)
    wfolder_4 = fileshare.create_directory(2, "4", wdir_1.dir_path, file_size)

    wdir_2 = fileshare.create_directory(0)
    wfolder_5 = fileshare.create_directory(2, "5", wdir_2.dir_path)

    files_to_transfer = [
        *wfolder_1.paths,
        wfolder_2.dir_path,
        wfolder_3.dir_path,
        wdir_1.dir_path,
        wdir_2.dir_path
        ]

    if background:
        output = sh.nordvpn.fileshare.send("--background", peer_address, *files_to_transfer).stdout.decode("utf-8")
        assert len(re.findall(r'File transfer ?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}) has started in the background.', output)) > 0
    else:
        command_handle = fileshare.start_transfer(peer_address, *files_to_transfer)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()

    for peer_transfer_id, _ in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if peer_transfer_id is not None:
            break

    assert peer_transfer_id is not None, "transfer was not received by peer"

    transfer_paths: list = \
        wfolder_1.filenames + \
        wfolder_2.transfer_paths + \
        wfolder_3.transfer_paths + \
        wfolder_4.transfer_paths + \
        wfolder_5.transfer_paths

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, transfer_paths, lambda file_entry: "request sent" in file_entry)

    peer_filepath = "/tmp/"
    if accept_entity == fileshare.FileSystemEntity.FILE:
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept --path {peer_filepath} {peer_transfer_id} {wfolder_1.filenames[1]}")

        transfer_paths.remove(wfolder_1.filenames[1])
        canceled_transfer_paths = transfer_paths

        transfer_local = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
        assert fileshare.for_all_files_in_transfer(transfer_local, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_local, [wfolder_1.filenames[1]], lambda file_entry: "uploaded" in file_entry)

        transfer_remote = ssh_client.exec_command(f"nordvpn fileshare list {local_transfer_id}")
        assert fileshare.for_all_files_in_transfer(transfer_remote, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_remote, [wfolder_1.filenames[1]], lambda file_entry: "downloaded" in file_entry)

        assert fileshare.files_from_transfer_exist_in_filesystem(local_transfer_id, [wfolder_1], ssh_client)
    elif accept_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept --path {peer_filepath} {peer_transfer_id} {os.path.basename(wfolder_2.dir_path)}")

        [transfer_paths.remove(path) for path in wfolder_2.transfer_paths]
        canceled_transfer_paths = transfer_paths

        transfer_local = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
        assert fileshare.for_all_files_in_transfer(transfer_local, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_local, wfolder_2.transfer_paths, lambda file_entry: "uploaded" in file_entry)

        transfer_remote = ssh_client.exec_command(f"nordvpn fileshare list {local_transfer_id}")
        assert fileshare.for_all_files_in_transfer(transfer_remote, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_remote, wfolder_2.transfer_paths, lambda file_entry: "downloaded" in file_entry)

        assert fileshare.files_from_transfer_exist_in_filesystem(local_transfer_id, [wfolder_2], ssh_client)
    else:
        # Directory
        t_progress_interactive = ssh_client.exec_command(f"nordvpn fileshare accept --path {peer_filepath} {peer_transfer_id} {os.path.basename(wdir_1.dir_path)}")

        [transfer_paths.remove(path) for path in wfolder_4.transfer_paths]
        canceled_transfer_paths = transfer_paths

        transfer_local = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
        assert fileshare.for_all_files_in_transfer(transfer_local, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_local, wfolder_4.transfer_paths, lambda file_entry: "uploaded" in file_entry)

        transfer_remote = ssh_client.exec_command(f"nordvpn fileshare list {local_transfer_id}")
        assert fileshare.for_all_files_in_transfer(transfer_remote, canceled_transfer_paths, lambda file_entry: "canceled" in file_entry)
        assert fileshare.for_all_files_in_transfer(transfer_remote, wfolder_4.transfer_paths, lambda file_entry: "downloaded" in file_entry)

        assert fileshare.files_from_transfer_exist_in_filesystem(local_transfer_id, [wfolder_4], ssh_client)

    time.sleep(1)

    transfers = sh.nordvpn.fileshare.list()
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    if not background:
        assert fileshare.validate_transfer_progress(command_handle.stdout.decode())
        assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, command_handle.stdout.decode())) == 1
        assert command_handle.is_alive() is False
        assert command_handle.exit_code == 0

    assert fileshare.validate_transfer_progress(t_progress_interactive)
    assert len(re.findall(fileshare.INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, t_progress_interactive)) == 1

    for entity in [wfolder_1, wfolder_2, wfolder_3, wdir_1, wdir_2]:
        shutil.rmtree(entity.dir_path)

    ssh_client.exec_command(f"sudo rm -rf {peer_filepath}/*tmp*")


@pytest.mark.parametrize("transfer_entity", list(fileshare.FileSystemEntity), ids = [f"send_{entity.value}" for entity in list(fileshare.FileSystemEntity)])
def test_fileshare_graceful_cancel(transfer_entity: fileshare.FileSystemEntity):
    wdir = fileshare.create_directory(0)
    wfolder = fileshare.create_directory(2, parent_dir=wdir.dir_path)

    if transfer_entity == fileshare.FileSystemEntity.FILE:
        path = wfolder.paths[0]
        expected_files = [wfolder.filenames[0]]
    elif transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        wfolder = fileshare.create_directory(2)
        path = wfolder.dir_path
        expected_files = wfolder.transfer_paths
    elif transfer_entity == fileshare.FileSystemEntity.DIRECTORY_WITH_FOLDERS:
        path = wdir.dir_path
        expected_files = wfolder.transfer_paths
    else: # fileshare.FileSystemEntity.FILES
        path = wfolder.paths
        expected_files = wfolder.filenames

    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    if transfer_entity == fileshare.FileSystemEntity.FILES:
        command_handle = fileshare.start_transfer(peer_address, *path)
    else:
        command_handle = fileshare.start_transfer(peer_address, path)

    for transfer_id, _ in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if transfer_id is not None:
            break

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "request sent" in fileshare.find_transfer_by_id(transfers_local, transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "waiting for download" in fileshare.find_transfer_by_id(transfers_remote, transfer_id)

    time.sleep(2)

    sh.kill("-s", "2", command_handle.pid)
    assert fileshare.MSG_CANCEL_TRANSFER in command_handle.stdout.decode()

    time.sleep(2)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id)
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "canceled by peer" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "canceled" in fileshare.find_transfer_by_id(transfers, local_transfer_id)

    if transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        shutil.rmtree(wfolder.dir_path)
    shutil.rmtree(wdir.dir_path)


@pytest.mark.parametrize("sender_cancels", [False, True], ids=["receiver_cancels", "sender_cancels"])
@pytest.mark.parametrize("transfer_entity", list(fileshare.FileSystemEntity), ids = [f"send_{entity.value}" for entity in list(fileshare.FileSystemEntity)])
def test_fileshare_graceful_cancel_transfer_ongoing(sender_cancels: bool, transfer_entity: fileshare.FileSystemEntity):
    file_size = 256
    wdir = fileshare.create_directory(0)
    wfolder = fileshare.create_directory(2, parent_dir=wdir.dir_path, file_size=file_size)

    if transfer_entity == fileshare.FileSystemEntity.FILE:
        path = wfolder.paths[0]
        expected_files = [wfolder.filenames[0]]
    elif transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        wfolder = fileshare.create_directory(2, file_size=file_size)
        path = wfolder.dir_path
        expected_files = wfolder.transfer_paths
    elif transfer_entity == fileshare.FileSystemEntity.DIRECTORY_WITH_FOLDERS:
        path = wdir.dir_path
        expected_files = wfolder.transfer_paths
    else: # fileshare.FileSystemEntity.FILES
        path = wfolder.paths
        expected_files = wfolder.filenames

    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    if transfer_entity == fileshare.FileSystemEntity.FILES:
        command_handle = fileshare.start_transfer(peer_address, *path)
    else:
        command_handle = fileshare.start_transfer(peer_address, path)

    for transfer_id, _ in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if transfer_id is not None:
            break

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "request sent" in fileshare.find_transfer_by_id(transfers_local, transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "waiting for download" in fileshare.find_transfer_by_id(transfers_remote, transfer_id)

    class PeerTransferAcceptThread(threading.Thread):
        def __init__(self, transfer_id: str):
            threading.Thread.__init__(self)
            self.message: str = ""

            self.transfer_id: str = transfer_id

        def run(self):
            self.message = ssh_client.exec_command(f"nordvpn fileshare accept {self.transfer_id}")

    transfer_accept_thread = PeerTransferAcceptThread(transfer_id)
    transfer_accept_thread.start()

    for transfer_in_progress in poll(lambda: "downloading" in fileshare.get_transfer(transfer_id, ssh_client)):
        if transfer_in_progress is not None:
            break

    assert transfer_in_progress, "transfer is either not started or already completed: " + str(fileshare.get_transfer(transfer_id, ssh_client))

    time.sleep(1) # avoid "downloading 0%"

    if sender_cancels:
        sh.kill("-s", "2", command_handle.pid)

        transfer_accept_thread.join()

        assert fileshare.MSG_CANCEL_TRANSFER in command_handle.stdout.decode()
        assert fileshare.validate_transfer_progress(command_handle.stdout.decode())
    else:
        fileshare_pid = ssh_client.exec_command("pgrep -f 'nordvpn fileshare accept'")
        ssh_client.exec_command(f"kill -s 2 {fileshare_pid}")

        transfer_accept_thread.join()

        assert fileshare.MSG_CANCEL_TRANSFER in transfer_accept_thread.message
        assert fileshare.validate_transfer_progress(transfer_accept_thread.message)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id)
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers_local = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    if sender_cancels:
        assert "canceled" in fileshare.find_transfer_by_id(transfers_local, local_transfer_id)
        assert "canceled by peer" in fileshare.find_transfer_by_id(transfers_remote, peer_transfer_id)
    else:
        assert "canceled by peer" in fileshare.find_transfer_by_id(transfers_local, peer_transfer_id)
        assert "canceled" in fileshare.find_transfer_by_id(transfers_remote, local_transfer_id)

    if transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        shutil.rmtree(wfolder.dir_path)
    shutil.rmtree(wdir.dir_path)


@pytest.mark.parametrize("background", [False, True], ids=["send_int", "send_bg"])
@pytest.mark.parametrize("sender_cancels", [False, True], ids=["receiver_cancels", "sender_cancels"])
@pytest.mark.parametrize("transfer_entity", list(fileshare.FileSystemEntity), ids = [f"send_{entity.value}" for entity in list(fileshare.FileSystemEntity)])
def test_fileshare_cancel_transfer(background: bool, transfer_entity: bool, sender_cancels: bool):
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    wdir = fileshare.create_directory(0)
    wfolder = fileshare.create_directory(2, parent_dir=wdir.dir_path)

    if transfer_entity == fileshare.FileSystemEntity.FILE:
        path = wfolder.paths[0]
        expected_files = [wfolder.filenames[0]]
    elif transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        wfolder = fileshare.create_directory(2)
        path = wfolder.dir_path
        expected_files = wfolder.transfer_paths
    elif transfer_entity == fileshare.FileSystemEntity.DIRECTORY_WITH_FOLDERS:
        path = wdir.dir_path
        expected_files = wfolder.transfer_paths
    else: # fileshare.FileSystemEntity.FILES
        path = wfolder.paths
        expected_files = wfolder.filenames

    if background:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, path)
    elif not background and transfer_entity == fileshare.FileSystemEntity.FILES:
        command_handle = fileshare.start_transfer(peer_address, *path)
    else:
        command_handle = fileshare.start_transfer(peer_address, path)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    transfers_local = sh.nordvpn.fileshare.list(_tty_out=False)
    assert "request sent" in fileshare.find_transfer_by_id(transfers_local, local_transfer_id)

    transfers_remote = ssh_client.exec_command("nordvpn fileshare list")
    assert "waiting for download" in fileshare.find_transfer_by_id(transfers_remote, peer_transfer_id)

    if sender_cancels:
        output = sh.nordvpn.fileshare.cancel(local_transfer_id)
        assert fileshare.CANCEL_SUCCESS_SENDER_SIDE_MSG in output
        if not background:
            output = command_handle.stdout.decode("utf-8")
            assert len(re.findall(fileshare.SEND_CANCELED_BY_OTHER_PROCESS_PATTERN, output)) > 0
    else: # receiver cancels
        output = ssh_client.exec_command(f"nordvpn fileshare cancel {peer_transfer_id}")
        assert fileshare.CANCEL_SUCCESS_SENDER_SIDE_MSG in output
        if not background:
            output = command_handle.stdout.decode("utf-8")
            assert len(re.findall(fileshare.SEND_CANCELED_BY_PEER_PATTERN, output)) > 0

    time.sleep(1)

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfer = sh.nordvpn.fileshare.list(local_transfer_id)
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "canceled" in file_entry)

    if sender_cancels:
        transfers = ssh_client.exec_command("nordvpn fileshare list")
        assert "canceled by peer" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
        assert "canceled" in fileshare.find_transfer_by_id(transfers, local_transfer_id)
    else:  # receiver cancels
        transfers = ssh_client.exec_command("nordvpn fileshare list")
        assert "canceled" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
        assert "canceled by peer" in fileshare.find_transfer_by_id(transfers, local_transfer_id)

    if transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        shutil.rmtree(wfolder.dir_path)
    shutil.rmtree(wdir.dir_path)


@pytest.mark.parametrize("sender_cancels", [True, False], ids=["sender_cancels", "receiver_cancels"])
def test_fileshare_cancel_file_not_in_flight(sender_cancels: bool):
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    wdir = fileshare.create_directory(5)
    sh.nordvpn.fileshare.send("--background", peer_address, wdir.dir_path)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    file_to_cancel = wdir.transfer_paths[2]
    if sender_cancels:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.fileshare.cancel(local_transfer_id, file_to_cancel)
            assert "This file is not in progress." in ex
            sh.nordvpn.fileshare.cancel(local_transfer_id)
    else:  # receiver cancels
        with pytest.raises(RuntimeError) as ex:
            ssh_client.exec_command(f"nordvpn fileshare cancel {peer_transfer_id} {file_to_cancel}")
            assert "This file is not in progress." in ex
            ssh_client.exec_command(f"nordvpn fileshare cancel {peer_transfer_id}")

    shutil.rmtree(wdir.dir_path)

@pytest.mark.parametrize("multiple_directories", [True, False], ids=["single_dir", "multi_dir"])
@pytest.mark.parametrize("background", [True, False], ids=["send_bg", "send_int"])
def test_fileshare_file_limit_exceeded(background: bool, multiple_directories: bool):
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    if not multiple_directories:
        wdir = fileshare.create_directory(1001)
        dirs = [wdir.dir_path]
    else:
        dir1 = fileshare.create_directory(400)
        dir2 = fileshare.create_directory(400)
        dir3 = fileshare.create_directory(201)
        dirs = [dir1.dir_path, dir2.dir_path, dir3.dir_path]

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if background:
            sh.nordvpn.fileshare.send("--background", peer_address, *dirs)
        else:
            sh.nordvpn.fileshare.send(peer_address, *dirs)

    assert "Number of files in a transfer cannot exceed 1000. Try archiving the directory." in str(ex.value)

    for directory_path in dirs:
        shutil.rmtree(directory_path)

@pytest.mark.parametrize("background", [True, False], ids=["send_bg", "send_int"])
def test_fileshare_file_directory_depth_exceeded(background: bool):
    src_path = tempfile.mkdtemp()

    path = src_path
    for _ in range(5):
        path = tempfile.mkdtemp(dir=path)

    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if background:
            sh.nordvpn.fileshare.send("--background", peer_address, src_path)
        else:
            sh.nordvpn.fileshare.send(peer_address, src_path)

    assert "File depth cannot exceed 5 directories. Try archiving the directory." in str(ex.value)

    shutil.rmtree(src_path)

def test_transfers_persistence():
    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer()

    sh.nordvpn.fileshare.send("--background", peer.ip, f"{workdir}/{test_files[0]}")

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    ssh_client.exec_command(f"nordvpn fileshare accept --path {workdir} {peer_transfer_id}")

    sh.nordvpn.set.mesh.off()
    sh.nordvpn.set.mesh.on()
    time.sleep(1)

    assert local_transfer_id in sh.nordvpn.fileshare.list()
    assert meshnet.is_peer_reachable(ssh_client, peer)  # Wait to reestablish connection for further tests
    sh.nordvpn.mesh.peer.refresh()


@pytest.mark.skip("load test, to be moved into a separate job")
def test_transfers_persistence_load():
    nordvpnd_pid = subprocess.Popen("/usr/bin/pidof nordvpnd", shell=True, stdout=subprocess.PIPE).stdout.read().strip()
    nordfileshared_pid = subprocess.Popen("/usr/bin/pidof nordfileshared", shell=True, stdout=subprocess.PIPE).stdout.read().strip()

    # multiple nordfileshared processes are found, choose only valid one
    pids = nordfileshared_pid.split()
    nordfileshared_pid = ""
    for pid in pids:
        # process should have valid parent
        ppid = sh.ps("-o", "ppid=", "-p", pid).strip()
        if len(sh.ps("-hp", ppid)) > 0:
            nordfileshared_pid = pid
            break

    assert len(nordfileshared_pid) > 0

    nordvpnd_memory_usage = get_memory_usage(int(nordvpnd_pid))
    nordfileshared_memory_usage = get_memory_usage(int(nordfileshared_pid))

    filepath = f"{workdir}/{test_files[0]}"

    # give time to start and connect
    time.sleep(1)

    peer = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer()
    assert len(peer.ip.strip()) != 0
    assert meshnet.is_peer_reachable(ssh_client, peer)

    min_send_time_ns = 100000000000  # 100s
    min_send_time_itr = 0
    max_send_time_ns = 0
    max_send_time_itr = 0
    sleep_s = 1
    sleep_ns = sleep_s * 1000 * 1000 * 1000  # nanoseconds

    transfers_count = 50
    nordfileshared_mem_delta_max_limit = 35000000  # bytes

    logging.log(f"test_transfers_persistence_load: start loop range: 0..{transfers_count}")

    total_send_time_ns = 0
    start_time_ns = time.time_ns()

    for i in range(transfers_count):
        start2_time_ns = time.time_ns()
        logging.log(f"[{i:5d}/{transfers_count}]a peer_address[ {peer.ip} ] file[ {filepath} ]")
        output = sh.nordvpn.fileshare.send("--background", peer.ip, filepath)
        logging.log(f"[{i:5d}/{transfers_count}]b fileshare send output[ {output} ]")
        time.sleep(sleep_s)

        local_transfer_id = fileshare.get_last_transfer()
        peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)
        logging.log(data=f"[{i:5d}/{transfers_count}]c transferID[ {local_transfer_id} ]")
        ssh_client.exec_command(f"nordvpn fileshare accept {peer_transfer_id}")

        send_time_ns = (time.time_ns() - start2_time_ns) - sleep_ns
        total_send_time_ns += send_time_ns
        if send_time_ns > max_send_time_ns:
            max_send_time_ns = send_time_ns
            max_send_time_itr = i
        if send_time_ns < min_send_time_ns:
            min_send_time_ns = send_time_ns
            min_send_time_itr = i

    finish_time_ns = time.time_ns()
    avg_send_time_ns = total_send_time_ns / transfers_count

    nordvpnd_memory_usage2 = get_memory_usage(int(nordvpnd_pid))
    nordfileshared_memory_usage2 = get_memory_usage(int(nordfileshared_pid))

    nordvpnd_memory_usage_delta = nordvpnd_memory_usage2 - nordvpnd_memory_usage
    nordfileshared_memory_usage_delta = nordfileshared_memory_usage2 - nordfileshared_memory_usage

    nordvpnd_memory_usage_per_tr = nordvpnd_memory_usage_delta / transfers_count
    nordfileshared_memory_usage_per_tr = nordfileshared_memory_usage_delta / transfers_count

    result = ssh_client.exec_command("nordvpn fileshare list | grep -c completed")
    completed_count = int(result)

    logging.log("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
    logging.log(f"transfers_count: {transfers_count} / completed_count: {completed_count}")
    logging.log(f" before test nordvpnd mem: {format_memory_usage(nordvpnd_memory_usage)} ;; nordfileshared mem: {format_memory_usage(nordfileshared_memory_usage)}/raw:{nordfileshared_memory_usage}")
    logging.log(f"  after test nordvpnd mem: {format_memory_usage(nordvpnd_memory_usage2)} ;; nordfileshared mem: {format_memory_usage(nordfileshared_memory_usage2)}/raw:{nordfileshared_memory_usage2}")
    logging.log(f"       nordvpnd mem delta: {format_memory_usage(nordvpnd_memory_usage_delta)} ;; nordfileshared mem delta: {format_memory_usage(nordfileshared_memory_usage_delta)}/raw:{nordfileshared_memory_usage_delta}")
    logging.log(f"nordvpnd mem per transfer: {format_memory_usage(nordvpnd_memory_usage_per_tr)} ;; nordfileshared mem per transfer: {format_memory_usage(nordfileshared_memory_usage_per_tr)}")
    logging.log(f"            avg send time: {format_time(avg_send_time_ns)} ;; min send time: {format_time(min_send_time_ns)}/@{min_send_time_itr} ;; max send time: {format_time(max_send_time_ns)}/@{max_send_time_itr}")
    logging.log(f"               total time: {format_time(finish_time_ns - start_time_ns)} ;; total send time: {format_time(total_send_time_ns)} ")
    logging.log("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

    assert completed_count >= transfers_count
    assert nordfileshared_mem_delta_max_limit > nordfileshared_memory_usage_delta


def get_memory_usage(pid):
    process = psutil.Process(pid)
    return process.memory_info().rss


def format_memory_usage(rss):
    KB = 1024  # noqa: N806
    MB = 1024 * 1024  # noqa: N806
    GB = 1024 * 1024 * 1024  # noqa: N806

    if rss >= GB:
        return f'{round(rss / GB, 2)}gb'
    if rss >= MB:
        return f'{round(rss / MB, 2)}mb'
    if rss >= KB:
        return f'{round(rss / KB, 2)}kb'
    return f'{round(rss, 2)}b'


def format_time(nanoseconds):
    if nanoseconds < 1000:
        return f'{int(nanoseconds)}ns'
    if nanoseconds < 1000000:
        return f'{int(nanoseconds / 1000)}μs'
    if nanoseconds < 1000000000:
        return f'{int(nanoseconds / 1000000)}ms'
    return f'{int(nanoseconds / 1000000000)}s'


@pytest.mark.parametrize("background", [True, False], ids=["send_bg", "send_int"])
@pytest.mark.parametrize("peer_name", list(meshnet.PeerName)[:-1])
def test_permissions_send(peer_name, background):
    tester_data = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device()
    tester_address = tester_data.get_peer_name(peer_name)

    fileshare_denied_message = ssh_client.exec_command(f"nordvpn mesh peer fileshare deny {tester_address}")
    tester_hostname = tester_data.get_peer_name(meshnet.PeerName.Hostname)
    assert meshnet.MSG_PEER_FILESHARE_DENY_SUCCESS % tester_hostname in fileshare_denied_message

    qa_peer_permission = meshnet.PeerList.from_str(ssh_client.exec_command("nordvpn mesh peer list")).get_internal_peer().allow_sending_files
    tester_permission = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().allows_sending_files
    assert tester_permission is not True
    assert qa_peer_permission is not True

    directory = fileshare.create_directory(1)
    filename = directory.paths[0]

    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().get_peer_name(peer_name)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if background:
            sh.nordvpn.fileshare.send("--background", peer_address, filename).stdout.decode("utf-8")
        else:
            sh.nordvpn.fileshare.send(peer_address, filename).stdout.decode("utf-8")

    assert "This peer does not allow file transfers from you." in str(ex.value)

    # Revert to the state before test
    fileshare_allowed_message = ssh_client.exec_command(f"nordvpn mesh peer fileshare allow {tester_address}")
    assert meshnet.MSG_PEER_FILESHARE_ALLOW_SUCCESS % tester_hostname in fileshare_allowed_message

    qapeer_permission = meshnet.PeerList.from_str(ssh_client.exec_command("nordvpn mesh peer list")).get_internal_peer().allow_sending_files
    tester_permission = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().allows_sending_files
    assert tester_permission is True
    assert qapeer_permission is True

    fileshare.start_transfer(peer_address, filename)
    transfer_id = fileshare.get_last_transfer()
    sh.nordvpn.fileshare.cancel(transfer_id)

    shutil.rmtree(directory.dir_path)


@pytest.mark.parametrize("peer_name", list(meshnet.PeerName)[:-1])
def test_permissions_meshnet_receive_forbidden(peer_name):
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().get_peer_name(peer_name)

    sh.nordvpn.mesh.peer.fileshare.deny(peer_address, _ok_code=[0, 1]).stdout.decode("utf-8")

    # transfer list should not change if transfer request was properly blocked
    expected_transfer_list = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    expected_transfer_list = expected_transfer_list[expected_transfer_list.index("Incoming"):].strip()

    tester_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().get_peer_name(peer_name)

    file_name = "/tmp/file_allowed"
    ssh_client.exec_command(f"echo > {file_name}")
    with pytest.raises(RuntimeError) as ex:
        ssh_client.exec_command(f"nordvpn fileshare send --background {tester_address} {file_name}")
        assert "peer does not allow file transfers" in ex

    ssh_client.exec_command(f"rm -rf {file_name}")

    actual_transfer_list = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    actual_transfer_list = actual_transfer_list[actual_transfer_list.index("Incoming"):].strip()

    assert expected_transfer_list == actual_transfer_list

    sh.nordvpn.mesh.peer.fileshare.allow(peer_address, _ok_code=[0, 1]).stdout.decode("utf-8")


def test_accept_destination_directory_does_not_exist():
    address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().ip

    filename = "file"

    ssh_client.exec_command(f"touch {workdir}/{filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {workdir}/{filename}")

    local_transfer_id = None
    error_message = None
    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):  # noqa: B007
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.fileshare.accept("--background", "--path", "invalid_dir", local_transfer_id).stdout.decode("utf-8")
        assert "Download directory invalid_dir does not exist. Make sure the directory exists or provide an alternative via --path" in ex

    sh.nordvpn.fileshare.cancel(local_transfer_id)


def test_accept_destination_directory_symlink():
    address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().ip

    filename = "file"

    ssh_client.exec_command(f"touch {workdir}/{filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {workdir}/{filename}")

    local_transfer_id = None
    error_message = None
    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):  # noqa: B007
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    dirpath = "/tmp/a"
    linkpath = "/tmp/b"

    os.mkdir(dirpath)
    os.symlink(dirpath, linkpath)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.fileshare.accept("--background", "--path", linkpath, local_transfer_id).stdout.decode("utf-8")
        assert f"Download directory {linkpath} is a symlink. You can provide provide an alternative via --path" in ex

    sh.nordvpn.fileshare.cancel(local_transfer_id)


def test_accept_destination_directory_not_a_directory():
    address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_this_device().ip

    filename = "file"

    ssh_client.exec_command(f"touch {workdir}/{filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {workdir}/{filename}")

    local_transfer_id = None
    error_message = None
    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):  # noqa: B007
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    _, path = tempfile.mkstemp()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.fileshare.accept("--background", "--path", path, local_transfer_id).stdout.decode("utf-8")
        assert f"Download directory {path} is a symlink. You can provide provide an alternative via --path" in ex

    sh.nordvpn.fileshare.cancel(local_transfer_id)

    sh.rm(path)


@pytest.mark.parametrize("transfer_entity", list(fileshare.FileSystemEntity), ids = [f"send_{entity.value}" for entity in list(fileshare.FileSystemEntity)])
def test_autoaccept(transfer_entity: fileshare.FileSystemEntity):
    peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())

    peer_name = peer_list.get_internal_peer().ip
    peer_hostname = peer_list.get_internal_peer().hostname
    msg = sh.nordvpn.mesh.peer("auto-accept", "enable", peer_name)
    assert meshnet.MSG_PEER_AUTOACCEPT_ALLOW_SUCCESS % peer_hostname in msg
    assert meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().accept_fileshare_automatically is True

    host_address = peer_list.get_this_device().ip

    wdir = fileshare.create_directory(0, ssh_client=ssh_client)
    wfolder = fileshare.create_directory(2, parent_dir=wdir.dir_path, ssh_client=ssh_client)

    if transfer_entity == fileshare.FileSystemEntity.FILE:
        path = wfolder.paths[0]
        expected_files = [wfolder.filenames[0]]
    elif transfer_entity == fileshare.FileSystemEntity.FOLDER_WITH_FILES:
        wfolder = fileshare.create_directory(2, ssh_client=ssh_client)
        path = wfolder.dir_path
        expected_files = wfolder.transfer_paths
    elif transfer_entity == fileshare.FileSystemEntity.DIRECTORY_WITH_FOLDERS:
        path = wdir.dir_path
        expected_files = wfolder.transfer_paths
    else: # fileshare.FileSystemEntity.FILES
        path = " ".join(wfolder.paths)
        expected_files = wfolder.filenames

    send_message = ssh_client.exec_command(f"nordvpn fileshare send --background {host_address} {path}")

    transfer_id = re.findall(f"{fileshare.TRANSFER_ID_REGEX}", send_message)[0]

    for peer_got_transfer in poll(lambda: (transfer_id in sh.nordvpn.fileshare.list())):
        if peer_got_transfer:
            break

    assert peer_got_transfer, "transfer was not received by peer"

    transfer = sh.nordvpn.fileshare.list(transfer_id)
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "downloaded" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, expected_files, lambda file_entry: "uploaded" in file_entry)

    assert fileshare.files_from_transfer_exist_in_filesystem(transfer_id, [wfolder])

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, transfer_id)

    msg = sh.nordvpn.mesh.peer("auto-accept", "disable", peer_name)
    assert meshnet.MSG_PEER_AUTOACCEPT_DENY_SUCCESS % peer_hostname in msg
    assert meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().accept_fileshare_automatically is False

    shutil.rmtree('/home/qa/Downloads')
    os.system(f"sudo mkdir -p -m 0777 {default_download_directory}")
    ssh_client.exec_command(f"rm -rf {wdir.dir_path}")


def test_peers_autocomplete():
    peer_hostname = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().hostname
    output = sh.nordvpn.fileshare.send("--generate-bash-completion")
    assert peer_hostname in output
    output = sh.nordvpn.fileshare.send(peer_hostname, "--generate-bash-completion")
    assert peer_hostname not in output  # Only first argument should be autocompleted


def test_transfers_autocomplete():
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip
    wdir = fileshare.create_directory(2)
    fileshare.start_transfer(peer_address, *wdir.paths)

    for peer_transfer_id, error_message in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client)):  # noqa: B007
        if peer_transfer_id is not None:
            break

    assert peer_transfer_id is not None, error_message

    print(ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}"))

    # Path should use default autocomplete
    output = ssh_client.exec_command("nordvpn fileshare accept --path --generate-bash-completion")
    assert peer_transfer_id not in output

    # Autocomplete transfers
    output = ssh_client.exec_command("nordvpn fileshare accept --path /tmp --generate-bash-completion")
    assert peer_transfer_id in output
    output = ssh_client.exec_command("nordvpn fileshare cancel --generate-bash-completion")
    assert peer_transfer_id in output
    output = ssh_client.exec_command("nordvpn fileshare list --generate-bash-completion")
    assert peer_transfer_id in output

    # Autocomplete transfer files
    output = ssh_client.exec_command(f"nordvpn fileshare accept --path /tmp {peer_transfer_id} --generate-bash-completion")
    assert wdir.filenames[0] in output
    assert wdir.filenames[1] in output

    ssh_client.exec_command(f"nordvpn fileshare accept --path /tmp {peer_transfer_id}")

    # When transfer is finished it should only be autocompleted in 'list' command
    output = ssh_client.exec_command("nordvpn fileshare accept --generate-bash-completion")
    assert peer_transfer_id not in output
    output = ssh_client.exec_command("nordvpn fileshare cancel --generate-bash-completion")
    assert peer_transfer_id not in output
    output = ssh_client.exec_command("nordvpn fileshare list --generate-bash-completion")
    assert peer_transfer_id in output

    shutil.rmtree(wdir.dir_path)

    for file in wdir.filenames:
        ssh_client.exec_command(f"rm -rf /tmp/{file}")


def test_clear():
    peer_address = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list()).get_internal_peer().ip

    sh.nordvpn.fileshare.send("--background", peer_address, f"{workdir}/{test_files[0]}")
    time.sleep(1)
    local_transfer_id0 = fileshare.get_last_transfer()

    for peer_transfer_id0 in poll(lambda: fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)):
        if peer_transfer_id0:
            break

    ssh_client.exec_command(f"nordvpn fileshare accept --path {workdir} {peer_transfer_id0}")

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id0)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id0)

    transfer_time0 = time.time()
    time.sleep(3)

    sh.nordvpn.fileshare.send("--background", peer_address, f"{workdir}/{test_files[1]}")
    time.sleep(1)
    local_transfer_id1 = fileshare.get_last_transfer()
    peer_transfer_id1 = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)
    ssh_client.exec_command(f"nordvpn fileshare accept --path {workdir} {peer_transfer_id1}")

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id0)
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id1)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id0)
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id1)

    fileshare.clear_history(f"{int(time.time() - transfer_time0)}")

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert fileshare.find_transfer_by_id(transfers, local_transfer_id0) is None
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id1)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id0)
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id1)

    fileshare.clear_history(f"{int(time.time() - transfer_time0)} seconds", ssh_client)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert fileshare.find_transfer_by_id(transfers, peer_transfer_id0) is None
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id1)

    time.sleep(3)
    fileshare.clear_history("1")

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert fileshare.find_transfer_by_id(transfers, local_transfer_id0) is None
    assert fileshare.find_transfer_by_id(transfers, local_transfer_id1) is None
    assert len(transfers.split("\n")) == 6

    lines_incoming = sh.nordvpn.fileshare.list("--incoming", _tty_out=False).split("\n")
    assert len(lines_incoming) == 3, str(lines_incoming)
    lines_outgoing = sh.nordvpn.fileshare.list("--outgoing", _tty_out=False).split("\n")
    assert len(lines_outgoing) == 3, str(lines_outgoing)

    fileshare.clear_history("all", ssh_client)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert fileshare.find_transfer_by_id(transfers, peer_transfer_id0) is None
    assert fileshare.find_transfer_by_id(transfers, peer_transfer_id1) is None
    assert len(transfers.split("\n")) == 6

    lines_incoming = ssh_client.exec_command("nordvpn fileshare list --incoming").split("\n")
    assert len(lines_incoming) == 3, str(lines_incoming)
    lines_outgoing = ssh_client.exec_command("nordvpn fileshare list --outgoing").split("\n")
    assert len(lines_outgoing) == 3, str(lines_outgoing)


def test_fileshare_process_monitoring_manages_fileshare_rules_on_process_state_changes():
    try:
        # port is open when fileshare is running
        assert fileshare.port_is_allowed()

        sh.pkill("-SIGKILL", "nordfileshare")
        # at the time of writing, the monitoring job is executed periodically every second,
        # wait for 2 seconds to be sure the job executed
        time.sleep(2)

        # port is not allowed when fileshare is down
        assert fileshare.port_is_blocked()

        # restart meshet to get fileshare back up
        fileshare.restart_mesh()

        # port is allowed again when fileshare process is up
        assert fileshare.port_is_allowed()
    finally: # meshnet should be on for most of the tests in this module
        fileshare.ensure_mesh_is_on()


@pytest.mark.skip(reason="LVPN-6691")
def test_fileshare_process_monitoring_cuts_the_port_access_even_when_it_was_taken_before():
    try:
        # stop meshnet to bind to 49111 first
        sh.nordvpn.set.meshnet.off()
        time.sleep(2)
        assert fileshare.port_is_blocked()

        # bind to port before fileshare process starts
        sock = fileshare.bind_port()
        assert sock is not None

        # start meshnet
        sh.nordvpn.set.meshnet.on() # now fileshare tries to start but fails because the port is taken
        time.sleep(2)

        # port should not be allowed (fileshare is down)
        assert fileshare.port_is_blocked()

        # free the port
        sock.close()

        # restart meshnet, now fileshare can start properly
        fileshare.restart_mesh()

        # fileshare is up so port is allowed
        assert fileshare.port_is_allowed()
    finally: # meshnet should be on for most of the tests in this module
        fileshare.ensure_mesh_is_on()


@pytest.mark.parametrize("background_accept", [True, False], ids=["accept_bg", "accept_int"])
@pytest.mark.parametrize("background_send", [True, False], ids=["send_bg", "send_int"])
def test_all_permissions_denied_send_file(background_send: bool, background_accept: bool):
    local_peer_list = meshnet.PeerList.from_str(sh.nordvpn.mesh.peer.list())
    local_address = local_peer_list.get_this_device().hostname

    permissions = ["incoming", "routing", "local"]

    for permission in permissions:
        with contextlib.suppress(RuntimeError):
            ssh_client.exec_command(f"nordvpn mesh peer {permission} deny {local_address}")

    wdir = fileshare.create_directory(1)
    peer_address = local_peer_list.get_internal_peer().hostname

    if background_send:
        sh.nordvpn.fileshare.send("--background", peer_address, wdir.paths[0])
    else:
        fileshare.start_transfer(peer_address, wdir.paths[0])

    remote_transfer_id = None
    error_message = None
    for remote_transfer_id, error_message in poll(lambda: fileshare.get_new_incoming_transfer(ssh_client)):  # noqa: B007
        if remote_transfer_id is not None:
            break

    assert remote_transfer_id is not None, error_message

    if background_accept:
        ssh_client.exec_command(f"nordvpn fileshare accept --background {remote_transfer_id}")
    else:
        ssh_client.exec_command(f"nordvpn fileshare accept {remote_transfer_id}")

    for transfers_done in poll(
        lambda: (
            "completed" in fileshare.get_transfer(remote_transfer_id) and
            "completed" in fileshare.get_transfer(remote_transfer_id, ssh_client)
        )
    ):
        if transfers_done:
            break

    peer_filepath = "~/Downloads/"
    assert fileshare.files_from_transfer_exist_in_filesystem(remote_transfer_id, [wdir], ssh_client)
    ssh_client.exec_command(f"rm -rf {peer_filepath}/{wdir.filenames[0]}")

    for permission in permissions:
        with contextlib.suppress(RuntimeError):
            ssh_client.exec_command(f"nordvpn mesh peer {permission} allow {local_address}")

    shutil.rmtree(wdir.dir_path)

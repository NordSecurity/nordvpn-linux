from lib import daemon, info, logging, login, meshnet, ssh, fileshare, poll
import lib
import sh
import re
import time
import pytest
import tempfile
import os
import psutil
import subprocess
import shutil
import pwd

import logging as logger

ssh_client = ssh.Ssh("qa-peer", "root", "root")

workdir = "/tmp"
testFiles = ["testing_fileshare_0.txt", "testing_fileshare_1.txt", "testing_fileshare_2.txt", "testing_fileshare_3.txt"]

default_download_directory = "/home/qa/Downloads"

def setup_module(module):
    daemon.start()
    login.login_as("default")
    lib.set_technology_and_protocol("nordlynx", "", "")

    # temporary hack for autoaccept tests, we create a default download directory
    # will be remove once default download directory setting is implemented
    os.system(f"sudo mkdir -m 0777 {default_download_directory}")

    sh.nordvpn.set.meshnet.on()
    # Ensure clean starting state
    meshnet.remove_all_peers()

    ssh_client.connect()
    daemon.install_peer(ssh_client)
    daemon.start_peer(ssh_client)
    login.login_as("default", ssh_client)
    ssh_client.exec_command("nordvpn set mesh on")

    sh.nordvpn.meshnet.peer.refresh()
    ssh_client.exec_command("nordvpn mesh peer refresh")
    assert meshnet.is_peer_reachable(ssh_client)

    message = "testing fileshare"
    for file in testFiles:
        filepath = f"{workdir}/{file}"
        with open(filepath, "w") as f:
            f.write(message)

    ssh_client.exec_command(f"mkdir /root/Downloads")


def teardown_module(module):
    dest_logs_path = f"{os.environ['CI_PROJECT_DIR']}/dist/logs"
    # Presere other peer log
    output = ssh_client.exec_command(f"cat /var/log/nordvpn/daemon.log")
    with open(f"{dest_logs_path}/other-peer-daemon.log", "w") as f:
        f.write(output)
    shutil.copy("/home/qa/.config/nordvpn/nordfileshared.log", dest_logs_path)
    ssh_client.exec_command("nordvpn set mesh off")
    ssh_client.exec_command("nordvpn logout --persist-token")
    daemon.stop_peer(ssh_client)
    daemon.uninstall_peer(ssh_client)
    ssh_client.disconnect()

    sh.nordvpn.set.meshnet.off()
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()
    ssh_client.exec_command("rm -rf /tmp/*")


@pytest.mark.parametrize("accept_directories",
    [["nested", "outer"],
    ["nested"],
    ["outer", "nested/inner"],
    ["nested/inner"]])
def test_accept(accept_directories):
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    # Check peer list on both ends
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

    ssh_client.exec_command(f"mkdir -p {nested_dir}")
    ssh_client.exec_command(f"echo > {nested_dir}/{filename}")
    ssh_client.exec_command(f"mkdir -p {nested_dir}/{inner_dir}")
    ssh_client.exec_command(f"echo > {nested_dir}/{inner_dir}/{filename}")
    ssh_client.exec_command(f"mkdir -p {outer_dir}")
    ssh_client.exec_command(f"echo > {outer_dir}/{filename}")

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
    output = ssh_client.exec_command(f"nordvpn fileshare send --background {address} {nested_dir} {outer_dir}")

    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    peer_transfer_id = fileshare.get_last_transfer(ssh_client=ssh_client)

    output = sh.nordvpn.fileshare.accept("--background", "--path", "/tmp", local_transfer_id, *accept_directories).stdout.decode("utf-8")

    def predicate(file_entry: str) -> bool:
        file_entry_columns = file_entry.split(' ')
        for directory in accept_directories:
            if file_entry_columns[0].startswith(directory) and ("downloaded" in file_entry or "uploaded" in file_entry):
                return True

        if "canceled" in file_entry:
            return True
        return False

    def check_files_status_receiver():
        transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
        return fileshare.for_all_files_in_transfer(transfer, transfer_files, predicate), transfer

    for receiver_files_status_ok, transfer in poll(check_files_status_receiver, attempts=10):
        if receiver_files_status_ok is True:
            break

    assert receiver_files_status_ok is True, f"invalid file status on receiver side, transfer {transfer}, files {accept_directories} should be downloaded"

    def check_files_status_sender():
        transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
        return fileshare.for_all_files_in_transfer(transfer, transfer_files, predicate), transfer

    for sender_files_status_ok, transfer in poll(check_files_status_sender, attempts=10):
        if sender_files_status_ok is True:
            break

    assert sender_files_status_ok is True, f"invalid file status on sender side, transfer {transfer}, files {accept_directories} should be uploaded"


@pytest.mark.parametrize("background", [True, False])
@pytest.mark.parametrize("peer_name", list(meshnet.PeerName))
def test_fileshare_transfer(background: bool, peer_name: meshnet.PeerName):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, peer_name)

    workdir = fileshare.create_directory(1)

    filepath = workdir.paths[0]
    message = "testing fileshare"
    with open(filepath, "w", encoding="utf-8") as file:
        file.write(message)

    if background:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, filepath)
        output = command_handle.stdout.decode("utf-8")
        assert len(re.findall(fileshare.SEND_NOWAIT_SUCCESS_MSG_PATTERN, output)) > 0
    else:
        command_handle = fileshare.start_transfer(peer_address, filepath)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = None

    for last_peer_transfer_id, _ in poll(lambda : fileshare.get_new_incoming_transfer(ssh_client), attempts=10):
        if last_peer_transfer_id is not None:
            peer_transfer_id = last_peer_transfer_id
            break

    assert peer_transfer_id is not None, "transfer was not received by peer"

    ssh_client.exec_command(f"nordvpn fileshare accept --path /tmp {peer_transfer_id}")

    time.sleep(1)

    peer_filepath = f"/tmp/{workdir.filenames[0]}"
    output = ssh_client.exec_command(f"cat {peer_filepath}")
    assert message in output

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert "downloaded" in transfer

    assert "uploaded" in sh.nordvpn.fileshare.list(local_transfer_id)

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id)


@pytest.mark.parametrize("background", [True, False])
@pytest.mark.parametrize("peer_name", list(meshnet.PeerName))
def test_fileshare_transfer_multiple_files(background: bool, peer_name: meshnet.PeerName):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, peer_name)

    dir1 = fileshare.create_directory(5, "1")
    dir2 = fileshare.create_directory(5, "2")
    dir3 = fileshare.create_directory(2, "3")

    # transfer dir1 and dir2 as directories and individual files from dir3, i.e /<dir1> /<dir2> /<dir3>/<file1> /<dir3>/<file2>
    files_to_transfer = [dir1.dir_path, dir2.dir_path] + dir3.paths

    if background:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, *files_to_transfer)
        output = command_handle.stdout.decode("utf-8")
        assert len(re.findall(fileshare.SEND_NOWAIT_SUCCESS_MSG_PATTERN, output)) > 0
    else:
        command_handle = fileshare.start_transfer(peer_address, *files_to_transfer)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    files_in_transfer = dir1.transfer_paths + dir2.transfer_paths + dir3.filenames

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "request sent" in file_entry)

    ssh_client.exec_command(f"nordvpn fileshare accept --path / {peer_transfer_id}")

    time.sleep(1)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "uploaded" in file_entry)

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert fileshare.for_all_files_in_transfer(transfer, files_in_transfer, lambda file_entry: "downloaded" in file_entry)

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers = ssh_client.exec_command("nordvpn fileshare list")
    assert "completed" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "completed" in fileshare.find_transfer_by_id(transfers, local_transfer_id)


@pytest.mark.parametrize("background", [True, False])
def test_fileshare_transfer_multiple_files_selective_accept(background: bool):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    workdir = fileshare.create_directory(4)

    if background:
        output = sh.nordvpn.fileshare.send("--background", peer_address, workdir.dir_path).stdout.decode("utf-8")
        assert len(re.findall(r'File transfer ?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}) has started in the background.', output)) > 0
    else:
        fileshare.start_transfer(peer_address, workdir.dir_path)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert fileshare.for_all_files_in_transfer(transfer, workdir.transfer_paths, lambda file_entry: "request sent" in file_entry)

    output = ssh_client.exec_command(
        f"nordvpn fileshare accept --path / {peer_transfer_id} {workdir.transfer_paths[0]} {workdir.transfer_paths[2]}"
    )

    time.sleep(1)

    transfer = sh.nordvpn.fileshare.list(local_transfer_id).stdout.decode("utf-8")
    assert "uploaded" in fileshare.find_file_in_transfer(workdir.transfer_paths[0], transfer.split("\n"))
    assert "canceled" in fileshare.find_file_in_transfer(workdir.transfer_paths[1], transfer.split("\n"))
    assert "uploaded" in fileshare.find_file_in_transfer(workdir.transfer_paths[2], transfer.split("\n"))
    assert "canceled" in fileshare.find_file_in_transfer(workdir.transfer_paths[3], transfer.split("\n"))

    transfer = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert "downloaded" in fileshare.find_file_in_transfer(workdir.transfer_paths[0], transfer.split("\n"))
    assert "canceled" in fileshare.find_file_in_transfer(workdir.transfer_paths[1], transfer.split("\n"))
    assert "downloaded" in fileshare.find_file_in_transfer(workdir.transfer_paths[2], transfer.split("\n"))
    assert "canceled" in fileshare.find_file_in_transfer(workdir.transfer_paths[3], transfer.split("\n"))


def test_fileshare_gracefull_cancel():
    workdir = fileshare.create_directory(1)

    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    command_handle = fileshare.start_transfer(peer_address, *workdir.paths)

    time.sleep(1)

    sh.kill("-s", "2", command_handle.pid)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    assert workdir.filenames[0] in sh.nordvpn.fileshare.list(local_transfer_id)
    assert "canceled" in sh.nordvpn.fileshare.list(local_transfer_id)

    output = ssh_client.exec_command(f"nordvpn fileshare list {peer_transfer_id}")
    assert workdir.filenames[0] in output
    assert "canceled" in output

    assert command_handle.is_alive() is False
    assert command_handle.exit_code == 0

    transfers = ssh_client.exec_command(f"nordvpn fileshare list")
    assert "canceled by peer" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

    transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    assert "canceled" in fileshare.find_transfer_by_id(transfers, local_transfer_id)


@pytest.mark.parametrize("background", [True, False])
@pytest.mark.parametrize("single_file", [True, False])
@pytest.mark.parametrize("sender_cancels", [True, False])
def test_fileshare_cancel_transfer(background: bool, single_file: bool, sender_cancels: bool):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    if single_file:
        dir = fileshare.create_directory(1)
        path = dir.paths[0]
        expected_files = [dir.filenames[0]]
    else:
        dir = fileshare.create_directory(5)
        path = dir.dir_path
        expected_files = dir.transfer_paths

    if background:
        command_handle = sh.nordvpn.fileshare.send("--background", peer_address, path)
    else:
        command_handle = fileshare.start_transfer(peer_address, path)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

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
    else: # receiver cancels
        transfers = ssh_client.exec_command("nordvpn fileshare list")
        assert "canceled" in fileshare.find_transfer_by_id(transfers, peer_transfer_id)

        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
        assert "canceled" in fileshare.find_transfer_by_id(transfers, local_transfer_id)


@pytest.mark.parametrize("sender_cancels", [True, False])
def test_fileshare_cancel_file_not_in_flight(sender_cancels: bool):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    workdir = fileshare.create_directory(5)
    sh.nordvpn.fileshare.send("--background", peer_address, workdir.dir_path)

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    file_to_cancel = workdir.transfer_paths[2]
    if sender_cancels:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            output = sh.nordvpn.fileshare.cancel(local_transfer_id, file_to_cancel)
            assert "This file is not in progress." in ex
    else: # receiver cancels
        with pytest.raises(RuntimeError) as ex:
            output = ssh_client.exec_command(f"nordvpn fileshare cancel {peer_transfer_id} {file_to_cancel}")
            assert "This file is not in progress." in ex


@pytest.mark.parametrize("background", [True, False])
@pytest.mark.parametrize("multiple_directories", [True, False])
def test_fileshare_file_limit_exceeded(background: bool, multiple_directories: bool):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    if multiple_directories:
        workdir = fileshare.create_directory(1001)
        dirs = [workdir.dir_path]
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


@pytest.mark.parametrize("background", [True, False])
def test_fileshare_file_directory_depth_exceeded(background: bool):
    src_path = tempfile.mkdtemp()

    path = src_path
    for _ in range(5):
        path = tempfile.mkdtemp(dir=path)

    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_this_device_ipv4(output)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        if background:
            sh.nordvpn.fileshare.send("--background", peer_address, src_path)
        else:
            sh.nordvpn.fileshare.send(peer_address, src_path)

    assert "File depth cannot exceed 5 directories. Try archiving the directory." in str(ex.value)

def test_transfers_persistence():
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_this_device_ipv4(output)

    sh.nordvpn.fileshare.send("--background", peer_address, f"{workdir}/{testFiles[0]}")

    time.sleep(1)

    local_transfer_id = fileshare.get_last_transfer()
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

    ssh_client.exec_command("nordvpn fileshare accept {}".format(peer_transfer_id))

    sh.nordvpn.set.mesh.off()
    sh.nordvpn.set.mesh.on()
    time.sleep(1)

    assert local_transfer_id in sh.nordvpn.fileshare.list()
    assert meshnet.is_peer_reachable(ssh_client) # Wait to reestablish connection for further tests
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

    filepath = f"{workdir}/{testFiles[0]}"

    # give time to start and connect
    time.sleep(1)

    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_this_device_ipv4(output)

    assert len(peer_address.strip()) != 0
    assert meshnet.is_peer_reachable(ssh_client)

    min_send_time_ns = 100000000000 # 100s
    min_send_time_itr = 0
    max_send_time_ns = 0
    max_send_time_itr = 0
    sleep_s = 1
    sleep_ns = sleep_s * 1000 * 1000 * 1000 # nanoseconds

    transfers_count = 50
    nordfileshared_mem_delta_max_limit = 35000000 # bytes

    logging.log("test_transfers_persistence_load: start loop range: 0..{}".format(transfers_count))

    total_send_time_ns = 0
    start_time_ns = time.time_ns()

    for i in range(transfers_count):
        start2_time_ns = time.time_ns()
        logging.log("[{:5d}/{}]a peer_address[ {} ] file[ {} ]".format(i, transfers_count, peer_address, filepath))
        output = sh.nordvpn.fileshare.send("--background", peer_address, filepath)
        logging.log("[{:5d}/{}]b fileshare send output[ {} ]".format(i, transfers_count, output))
        time.sleep(sleep_s)

        local_transfer_id = fileshare.get_last_transfer()
        peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)
        logging.log(data="[{:5d}/{}]c transferID[ {} ]".format(i, transfers_count, local_transfer_id))
        ssh_client.exec_command("nordvpn fileshare accept {}".format(peer_transfer_id))

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
    logging.log("transfers_count: {} / completed_count: {}".format(transfers_count, completed_count))
    logging.log(" before test nordvpnd mem: {} ;; nordfileshared mem: {}/raw:{}".format(format_memory_usage(nordvpnd_memory_usage), format_memory_usage(nordfileshared_memory_usage), nordfileshared_memory_usage))
    logging.log("  after test nordvpnd mem: {} ;; nordfileshared mem: {}/raw:{}".format(format_memory_usage(nordvpnd_memory_usage2), format_memory_usage(nordfileshared_memory_usage2), nordfileshared_memory_usage2))
    logging.log("       nordvpnd mem delta: {} ;; nordfileshared mem delta: {}/raw:{}".format(format_memory_usage(nordvpnd_memory_usage_delta), format_memory_usage(nordfileshared_memory_usage_delta), nordfileshared_memory_usage_delta))
    logging.log("nordvpnd mem per transfer: {} ;; nordfileshared mem per transfer: {}".format(format_memory_usage(nordvpnd_memory_usage_per_tr), format_memory_usage(nordfileshared_memory_usage_per_tr)))
    logging.log("            avg send time: {} ;; min send time: {}/@{} ;; max send time: {}/@{}".format(format_time(avg_send_time_ns), format_time(min_send_time_ns), min_send_time_itr, format_time(max_send_time_ns), max_send_time_itr))
    logging.log("               total time: {} ;; total send time: {} ".format(format_time(finish_time_ns-start_time_ns), format_time(total_send_time_ns)))
    logging.log("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

    assert completed_count >= transfers_count
    assert nordfileshared_mem_delta_max_limit > nordfileshared_memory_usage_delta


def get_memory_usage(pid):
    process = psutil.Process(pid)
    return process.memory_info().rss

def format_memory_usage(rss):
    KB = 1024
    MB = 1024 * 1024
    GB = 1024 * 1024 * 1024

    if rss >= GB:
        return f'{round(rss / GB, 2)}gb'
    elif rss >= MB:
        return f'{round(rss / MB, 2)}mb'
    elif rss >= KB:
        return f'{round(rss / KB, 2)}kb'
    else:
        return f'{round(rss, 2)}b'

def format_time(nanoseconds):
    if nanoseconds < 1000:
        return f'{int(nanoseconds)}ns'
    elif nanoseconds < 1000000:
        return f'{int(nanoseconds / 1000)}μs'
    elif nanoseconds < 1000000000:
        return f'{int(nanoseconds / 1000000)}ms'
    else:
        return f'{int(nanoseconds / 1000000000)}s'

@pytest.mark.parametrize("peer_name", list(meshnet.PeerName))
def test_permissions_send_allowed(peer_name):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, peer_name)

    directory = fileshare.create_directory(1)
    filename = directory.paths[0]

    output = sh.nordvpn.fileshare.send("--background", peer_address, filename).stdout.decode("utf-8")
    transfer_id = re.findall(fileshare.SEND_NOWAIT_SUCCESS_MSG_PATTERN, output)[0]
    assert transfer_id is not None

    time.sleep(1)

    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)
    assert peer_transfer_id is not None


@pytest.mark.parametrize("peer_name", list(meshnet.PeerName))
def test_permissions_send_forbidden(peer_name):
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    tester_address = meshnet.get_peer_name(output, peer_name)

    ssh_client.exec_command(f"nordvpn mesh peer fileshare deny {tester_address}")

    directory = fileshare.create_directory(1)
    filename = directory.paths[0]

    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, peer_name)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.fileshare.send("--background", peer_address, filename).stdout.decode("utf-8")

    assert "This peer does not allow file transfers from you." in str(ex.value)

    # Revert to the state before test
    ssh_client.exec_command(f"nordvpn mesh peer fileshare allow {tester_address}")


@pytest.mark.parametrize("peer_name", list(meshnet.PeerName))
def test_permissions_meshnet_receive_forbidden(peer_name):
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_this_device_ipv4(output)

    sh.nordvpn.mesh.peer.fileshare.deny(peer_address, _ok_code=[0, 1]).stdout.decode("utf-8")

    # transfer list should not change if transfer request was properly blocked
    expected_transfer_list = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    expected_transfer_list = expected_transfer_list[expected_transfer_list.index("Incoming"):].strip()

    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    tester_addess = meshnet.get_peer_name(output, peer_name)

    file_name = "/tmp/file_allowed"
    ssh_client.exec_command(f"echo > {file_name}")
    with pytest.raises(RuntimeError) as ex:
        ssh_client.exec_command(f"nordvpn fileshare send --background {tester_addess} {file_name}")
        assert "peer does not allow file transfers" in ex

    actual_transfer_list = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    actual_transfer_list = actual_transfer_list[actual_transfer_list.index("Incoming"):].strip()

    assert expected_transfer_list == actual_transfer_list

    sh.nordvpn.mesh.peer.fileshare.allow(peer_address, _ok_code=[0, 1]).stdout.decode("utf-8")


def test_accept_destination_directory_does_not_exist():
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    filename = "file"

    ssh_client.exec_command(f"touch {filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {filename}")

    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        output = sh.nordvpn.fileshare.accept("--background", "--path", "invalid_dir", local_transfer_id).stdout.decode("utf-8")
        assert "Download directory invalid_dir does not exist. Make sure the directory exists or provide an alternative via --path" in ex


def test_accept_destination_directory_symlink():
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    filename = "file"

    ssh_client.exec_command(f"touch {filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {filename}")

    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    dirpath = "/tmp/a"
    linkpath = "/tmp/b"

    os.mkdir(dirpath)
    os.symlink(dirpath, linkpath)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        output = sh.nordvpn.fileshare.accept("--background", "--path", linkpath, local_transfer_id).stdout.decode("utf-8")
        assert f"Download directory {linkpath} is a symlink. You can provide provide an alternative via --path" in ex


def test_accept_destination_directory_not_a_directory():
    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    filename = "file"

    ssh_client.exec_command(f"touch {filename}")
    ssh_client.exec_command(f"nordvpn fileshare send --background {address} {filename}")

    for local_transfer_id, error_message in poll(fileshare.get_new_incoming_transfer):
        if local_transfer_id is not None:
            break

    assert local_transfer_id is not None, error_message

    _, path = tempfile.mkstemp()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        output = sh.nordvpn.fileshare.accept("--background", "--path", path, local_transfer_id).stdout.decode("utf-8")
        assert f"Download directory {path} is a symlink. You can provide provide an alternative via --path" in ex


def test_autoaccept():
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_name = meshnet.get_peer_name(output, meshnet.PeerName.Ip)
    output = subprocess.run(["nordvpn", "mesh", "peer", "auto-accept", "enable", peer_name])
    # subprocess.run(["nordvpn", "mesh", "peer", "auto-accept", "enable", peer_name])

    time.sleep(10)

    output = f'{sh.nordvpn.mesh.peer.list(_tty_out=False)}'
    host_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)

    filename = "autoaccepted"
    peer_file_path = f"/home/qapeer/{filename}"
    ssh_client.exec_command(f"echo > {peer_file_path}")
    output = ssh_client.exec_command(f"nordvpn fileshare send --background {host_address} {peer_file_path}")

    def check_if_file_received():
        last_transfer_id = fileshare.get_last_transfer(outgoing=False)
        transfer = sh.nordvpn.fileshare.list(last_transfer_id).stdout.decode("utf-8")
        transfer_lines = transfer.split("\n")
        file_entry = fileshare.find_file_in_transfer(filename, transfer_lines)
        return file_entry is not None and "downloaded" in file_entry

    for file_received in poll(check_if_file_received, attempts=10):
        if file_received is True:
            break

    assert file_received
    assert os.path.isfile(f"{default_download_directory}/{filename}")


def test_peers_autocomplete():
    peer_hostname = meshnet.get_this_device(ssh_client.exec_command("nordvpn mesh peer list"))
    output = sh.nordvpn.fileshare.send("--generate-bash-completion")
    assert peer_hostname in output
    output = sh.nordvpn.fileshare.send(peer_hostname, "--generate-bash-completion")
    assert peer_hostname not in output # Only first argument should be autocompleted


def test_transfers_autocomplete():
    output = ssh_client.exec_command("nordvpn mesh peer list")
    peer_address = meshnet.get_peer_name(output, meshnet.PeerName.Ip)
    workdir = fileshare.create_directory(2)
    fileshare.start_transfer(peer_address, *workdir.paths)

    time.sleep(1)
    peer_transfer_id = fileshare.get_last_transfer(outgoing=False, ssh_client=ssh_client)

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
    assert workdir.filenames[0] in output 
    assert workdir.filenames[1] in output

    ssh_client.exec_command(f"nordvpn fileshare accept --path /tmp {peer_transfer_id}")

    # When transfer is finished it should only be autocompleted in 'list' command
    output = ssh_client.exec_command("nordvpn fileshare accept --generate-bash-completion")
    assert peer_transfer_id not in output
    output = ssh_client.exec_command("nordvpn fileshare cancel --generate-bash-completion")
    assert peer_transfer_id not in output
    output = ssh_client.exec_command("nordvpn fileshare list --generate-bash-completion")
    assert peer_transfer_id in output

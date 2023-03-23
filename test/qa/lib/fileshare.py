from lib import ssh
from typing import Callable
from collections import namedtuple
import tempfile
import os
import sh
import re

SEND_NOWAIT_SUCCESS_MSG_PATTERN = r'File transfer ?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}) has started in the background.'
SEND_CANCELED_BY_PEER_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by peer'
SEND_CANCELED_BY_OTHER_PROCESS_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by other process'
CANCEL_SUCCESS_SENDER_SIDE_MSG = "File transfer canceled"

Directory = namedtuple("Directory", "dir_path paths transfer_paths filenames")

def create_directory(file_count: int, name_suffix: str = "", parent_dir: str = None) -> Directory:
    dir_path = tempfile.mkdtemp(dir=parent_dir)
    paths = []
    transfer_paths = []
    filenames = []

    for file_number in range(file_count):
        filename = f"file_{file_number}{name_suffix}"
        path = f"{dir_path}/{filename}"
        paths.append(path)
        # in transfer, files are displayed with leading directory only, i.e /tmp/dir/file becomes dir/file
        transfer_paths.append(path.removeprefix("/tmp/"))
        filenames.append(filename)
        os.mknod(f"{dir_path}/{filename}")

    return Directory(dir_path, paths, transfer_paths, filenames)


def start_transfer(peer_address: str, *filepaths: str) -> sh.RunningCommand:
    command = sh.nordvpn.fileshare.send(peer_address, filepaths, _iter=True, _out_bufsize=0)
    buffer = ""

    # Read the output character by character, we cannot read it line by line because send prints some
    # messages without the newline at the end and we cannot read it by X characters because before command
    # is executed app connets to the daemon for non-deterministic ammount of time displaying a spinner
    for character in command:
        buffer += character
        if "Waiting for the peer to accept your transfer..." in buffer:
            break

    return command


def get_last_transfer(outgoing: bool = True, ssh_client: ssh.Ssh = None) -> str:
    if ssh_client is None:
        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    else:
        transfers = ssh_client.exec_command(f"nordvpn fileshare list")
    outgoing_index = transfers.index("Outgoing")
    transfers = transfers[outgoing_index:] if outgoing else transfers[:outgoing_index]
    transfer_ids = re.findall("([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})", transfers)

    if len(transfer_ids) == 0:
        return None

    return re.findall("([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})", transfers)[-1]


def get_transfer(transfer_id: str, ssh_client: ssh.Ssh = None) -> str:
    if ssh_client is None:
        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    else:
        transfers = ssh_client.exec_command(f"nordvpn fileshare list")

    transfer_entry = [transfer_entry for transfer_entry in transfers.split("\n") if transfer_id in transfer_entry]

    if len(transfer_entry) == 0:
        return None

    return transfer_entry[0]


def find_transfer_by_id(transfer_list: str, id: str) -> str:
    for transfer_entry in transfer_list.strip("\n").split("\n"):
        if id in transfer_entry:
            return transfer_entry
    return None


def find_file_in_transfer(file_id: str, transfer_lines: list[str]) -> str:
    for file_entry in transfer_lines:
        if file_id in file_entry:
            return file_entry
    return None


# returns true if all of the files are present in the transfer and meet the predicate
def for_all_files_in_transfer(transfer: str, files: list[str], predicate: Callable[[str], bool]) -> bool:
    transfer = transfer.split("\n")
    for file in files:
        file_entry = find_file_in_transfer(file, transfer)
        if file_entry is None or not predicate(file_entry):
            return False
    return True


# returns last incoming transfer that has not completed
def get_new_incoming_transfer():
    local_transfer_id = get_last_transfer(outgoing=False)
    if local_transfer_id is None:
        return None, f"there are no started transfers"

    transfer_status = get_transfer(local_transfer_id)
    if transfer_status is None:
        return None, f"could not read transfer {local_transfer_id} status on receiver side after it has been initiated by the sender"
    if "completed" in transfer_status:
        return None, f"no new transfers found on receiver side after transfer has been initiated by the sender, last transfer is {local_transfer_id} but its status is completed"
    return local_transfer_id, ""
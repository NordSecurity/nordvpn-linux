import os
import re
import tempfile
from collections import namedtuple
from collections.abc import Callable

import sh

from . import logging, ssh

SEND_NOWAIT_SUCCESS_MSG_PATTERN = r'File transfer ?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}) has started in the background.'
SEND_CANCELED_BY_PEER_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by peer'
SEND_CANCELED_BY_OTHER_PROCESS_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by other process'
CANCEL_SUCCESS_SENDER_SIDE_MSG = "File transfer canceled"
TRANSFER_ID_REGEX = r"[a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}"

MSG_HISTORY_CLEARED = "File transfer history cleared."
MSG_CANCEL_TRANSFER = "File transfer canceled."

Directory = namedtuple("Directory", "dir_path paths transfer_paths filenames")


def create_directory(file_count: int, name_suffix: str = "", parent_dir: str | None = None, file_size: str = "1K") -> Directory:
    # for snap testing make directories to be created from current path e.g. dir="./"
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

        disallowed_filesize = ["G", "T", "P", "E", "Z", "Y"]
        for size in disallowed_filesize:
            if size in file_size:
                raise ValueError("Specified file size is too big. Specify either (K)ilobytes or (M)egabytes")

        sh.fallocate("-l", file_size, f"{dir_path}/{filename}")

    return Directory(dir_path, paths, transfer_paths, filenames)


def start_transfer(peer_address: str, *filepaths: str) -> sh.RunningCommand:
    """
    Initiates a file transfer to a specified peer.

    Args:
        peer_address (str): The address of the peer to send files to.
        *filepaths (str): One or more file paths to be transferred.

    Returns:
        sh.RunningCommand: The running command object, which allows interaction with the ongoing process.

    Further code execution is blocked, until this function finds "Waiting for the peer to accept your transfer..."
    string, indicating, that the transfer process has started and is waiting for peer confirmation.
    """
    command = sh.nordvpn.fileshare.send(peer_address, filepaths, _iter=True, _out_bufsize=0)
    buffer = ""

    # Read the output character by character, we cannot read it line by line because send prints some
    # messages without the newline at the end, and we cannot read it by X characters because before command
    # is executed app connects to the daemon for non-deterministic amount of time displaying a spinner
    for character in command:
        buffer += character
        if "Waiting for the peer to accept your transfer..." in buffer:
            break

    return command


def get_last_transfer(outgoing: bool = True, ssh_client: ssh.Ssh = None) -> str | None:
    """Return last id of the last received or sent transfer."""
    if ssh_client is None:
        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    else:
        transfers = ssh_client.exec_command("nordvpn fileshare list")
    outgoing_index = transfers.index("Outgoing")
    transfers = transfers[outgoing_index:] if outgoing else transfers[:outgoing_index]
    transfer_ids = re.findall(f"({TRANSFER_ID_REGEX})", transfers)

    if len(transfer_ids) == 0:
        return None

    return transfer_ids[-1]


def get_transfer(transfer_id: str, ssh_client: ssh.Ssh = None) -> str | None:
    if ssh_client is None:
        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    else:
        transfers = ssh_client.exec_command("nordvpn fileshare list")

    transfer_entry = [transfer_entry for transfer_entry in transfers.split("\n") if transfer_id in transfer_entry]

    if len(transfer_entry) == 0:
        return None

    return transfer_entry[0]


def find_transfer_by_id(transfer_list: str, idd: str) -> str | None:
    for transfer_entry in transfer_list.strip("\n").split("\n"):
        if idd in transfer_entry:
            return transfer_entry
    return None


def find_file_in_transfer(file_id: str, transfer_lines: list[str]) -> str | None:
    for file_entry in transfer_lines:
        if file_id in file_entry:
            return file_entry
    return None


# returns true if all the files are present in the transfer and meet the predicate
def for_all_files_in_transfer(transfer: str, files: list[str], predicate: Callable[[str], bool]) -> bool:
    transfer = transfer.split("\n")
    for file in files:
        file_entry = find_file_in_transfer(file, transfer)
        if file_entry is None or not predicate(file_entry):
            return False
    return True


def get_new_incoming_transfer(ssh_client: ssh.Ssh = None):
    """Returns last incoming transfer that has not completed."""
    local_transfer_id = get_last_transfer(outgoing=False, ssh_client=ssh_client)
    if local_transfer_id is None:
        return None, "there are no started transfers"

    transfer_status = get_transfer(local_transfer_id, ssh_client)
    if transfer_status is None:
        return None, f"could not read transfer {local_transfer_id} status on receiver side after it has been initiated by the sender"
    if "completed" in transfer_status:
        return None, f"no new transfers found on receiver side after transfer has been initiated by the sender, last transfer is {local_transfer_id} but its status is completed"
    return local_transfer_id, ""


def cancel_not_finished_transfers():
    transfers = get_not_finished_transfers()
    for transfer_id in transfers:
        try:
            sh.nordvpn.fileshare.cancel(transfer_id).stdout.decode("utf-8")
        except sh.ErrorReturnCode_1 as ex:
            logging.log(f"failed to cancel transfer {transfer_id}: {ex}")


def get_not_finished_transfers(ssh_client: ssh.Ssh = None) -> list[str]:
    """Return IDs of of all transfers which are not: completed or  canceled."""
    if ssh_client is None:
        transfers = sh.nordvpn.fileshare.list().stdout.decode("utf-8")
    else:
        transfers = ssh_client.exec_command("nordvpn fileshare list")
    # all transfer IDs without "completed" or "canceled" following the ID
    transfer_ids = re.findall(f"({TRANSFER_ID_REGEX})(?!.*(?:completed|canceled))", transfers)

    if len(transfer_ids) == 0:
        return []

    return transfer_ids


def clear_history(time_period: str, ssh_client: ssh.Ssh = None):
    """
    Clears the fileshare history for a specified time period, either locally or via SSH.

    Args:
        time_period (str): The time period for which to clear the fileshare history.
        ssh_client (ssh.Ssh, optional): SSH client for executing the command on a remote server.
                                        If None, executes locally.

    Raises:
        AssertionError: If the history clearing message is not found.
    """
    if ssh_client is None:
        msg = sh.nordvpn.fileshare.clear(time_period)
    else:
        msg = ssh_client.exec_command(f"nordvpn fileshare clear {time_period}")

    assert MSG_HISTORY_CLEARED in msg

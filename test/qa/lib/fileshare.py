import re
import tempfile
import time
from collections import namedtuple
from collections.abc import Callable
from enum import Enum
from threading import Thread

import pytest
import sh

from . import FILE_HASH_UTILITY, logging, ssh

SEND_NOWAIT_SUCCESS_MSG_PATTERN = r'File transfer ?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}) has started in the background.'
SEND_CANCELED_BY_PEER_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by peer'
SEND_CANCELED_BY_OTHER_PROCESS_PATTERN = r'File transfer \[?([a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12})\] canceled by other process'
CANCEL_SUCCESS_SENDER_SIDE_MSG = "File transfer canceled"
TRANSFER_ID_REGEX = r"[a-z0-9]{8}-(?:[a-z0-9]{4}-){3}[a-z0-9]{12}"
INTERACTIVE_TRANSFER_PROGRESS_ONGOING_PATTERN = r"File transfer \[[0-9a-fA-F\-]{36}\] progress \[(\d{1,3})%\]"
INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN = r"File transfer \[[0-9a-fA-F\-]{36}\] completed."

MSG_HISTORY_CLEARED = "File transfer history cleared."
MSG_CANCEL_TRANSFER = "File transfer canceled."

Directory = namedtuple("Directory", "dir_path paths transfer_paths filenames filehashes")


def create_directory(file_count: int, name_suffix: str = "", parent_dir: str | None = None, file_size: str = "1K") -> Directory:
    """
    Creates a temporary directory and populates it with a specified number of files.

    Args:
        file_count (int): The number of files to create in the directory.
        name_suffix (str, optional): A suffix to append to the filenames. Defaults to an empty string.
        parent_dir (str | None, optional): The parent directory where the temporary directory will be created.
                                           If None, the system default temporary directory is used. Defaults to None.
        file_size (str, optional): The size of each file to be created, specified using typical file size notation
                                   (e.g., "1K", "128M"). Defaults to "1K".
    Returns:
        Directory: A Directory object containing:
            - dir_path: Path to the created directory.
            - paths: Full paths to the created files.
            - transfer_paths: File paths with leading directories removed.
            - filenames: Names of the created files.
    """
    # for snap testing make directories to be created from current path e.g. dir="./"
    dir_path = tempfile.mkdtemp(dir=parent_dir)
    paths = []
    transfer_paths = []
    filenames = []
    filehashes = []

    hash_util = sh.Command(FILE_HASH_UTILITY)

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

        hash_output = hash_util(path).strip().split()[0]  # Only take the hash part of the output
        filehashes.append(hash_output)

    return Directory(dir_path, paths, transfer_paths, filenames, filehashes)


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
    """Returns last incoming transfer that is not finished."""
    local_transfer_id = get_last_transfer(outgoing=False, ssh_client=ssh_client)
    if local_transfer_id is None:
        return None, "there are no started transfers"

    transfer_status = get_transfer(local_transfer_id, ssh_client)
    if transfer_status is None:
        return None, f"could not read transfer {local_transfer_id} status on receiver side after it has been initiated by the sender"
    if "completed" in transfer_status or "canceled" in transfer_status:
        return None, f"no new transfers found on receiver side after transfer has been initiated by the sender, last transfer is {local_transfer_id} but its status is {transfer_status}"
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


def validate_transfer_progress(transfer_log: str):
    """
    Checks if transfer progress in the log is consistently increasing.

    Extracts transfer progress lines from the log, ensures they share the same
    transfer ID, and verifies that the progress percentage increases without decreasing.

    Args:
        transfer_log (str): The log containing transfer progress details.
    Returns:
        bool: True if progress is increasing, False if not.
    Raises:
        AssertionError: If message about completed transfer is missing
        AssertionError: If transfer ID is inconsistent.
    """
    assert len(re.findall(INTERACTIVE_TRANSFER_PROGRESS_COMPLETED_PATTERN, transfer_log)) == 1
    filtered_lines = [line for line in transfer_log.split("\r") if re.search(INTERACTIVE_TRANSFER_PROGRESS_ONGOING_PATTERN, line)]
    transfer_id = re.search(TRANSFER_ID_REGEX, filtered_lines[0])

    previous_progress = -1
    increasing = True

    for line in filtered_lines:
        assert transfer_id.group() in line

        precentage_match = re.findall(r"\[(\d{1,3})%\]", line)
        if len(precentage_match) == 1:
            precentage = int(precentage_match[0])
        else:
            pytest.fail(f"Precentage was not found during the validation of interactive transfer progress: {line}")

        if precentage < previous_progress:
            increasing = False
            break
        previous_progress = precentage

    return increasing


class TransferState(Enum):
    DOWNLOADING = "downloading"
    UPLOADING = "uploading"

    def __str__(self):
        return self.value

class TransferProgressValidationThread(Thread):
    def __init__(self, transfer_id: str, expected_state: str, ssh_client: ssh.Ssh = None):
        Thread.__init__(self)
        self.transfer_progress_valid: bool = False

        self.transfer_id: str = transfer_id
        self.ssh_client: ssh.Ssh = ssh_client
        self.expected_state: str = expected_state

    def run(self):
        self.transfer_progress_valid = validate_transfer_progress_bg(self.transfer_id, self.ssh_client, self.expected_state)

def validate_transfer_progress_bg(transfer_id: str, ssh_client: ssh.Ssh, expected_state: str) -> bool:
    """
    Checks if transfer progress in the `nordvpn fileshare list` is consistently increasing.

    Extracts transfer progress from `nordvpn fileshare list`, and
    verifies that the progress percentage increases without decreasing.

    Args:
        transfer_id (str): The transfer whose progress we want to track.
    Returns:
        bool: True if progress is increasing, False if not.
    """
    progress_log: list[int] = []
    previous_progress = -1
    increasing = True

    retry = 0

    while True:
        if ssh_client:
            transfers = ssh_client.exec_command("nordvpn fileshare list")
        else:
            transfers = sh.nordvpn.fileshare.list(_tty_out=False)

        transfer = find_transfer_by_id(transfers, transfer_id)

        if "completed" in transfer:
            break  # Exit the loop when the transfer is completed

        # Extract the percentage progress using regex
        matches = re.findall(f"{expected_state}" + r"\s+(\d{1,3})%", transfer)

        if len(matches) == 1:
            percentage = int(matches[0])
            progress_log.append(percentage)

            if percentage < previous_progress and previous_progress != -1:
                print(f"Progress log: {str(progress_log)}")
                increasing = False
                break  # Exit if the progress is decreasing

            previous_progress = percentage
        elif retry == 3:
            pytest.fail(f"Precentage was not found during the validation of interactive transfer progress:\n{transfer}")
        else:
            retry += 1
            time.sleep(1)
    return increasing

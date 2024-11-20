import contextlib
import json
import time
from collections import namedtuple

import paramiko
import pytest

import lib

Directory = namedtuple("Directory", "dir_path paths transfer_paths filenames filehashes")


class Ssh:
    def __init__(self, hostname: str, username: str, password: str):
        self.client = paramiko.SSHClient()
        self.hostname = hostname
        self.username = username
        self.password = password
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        self.daemon = self.Daemon(self)
        self.io = self.IO(self)
        self.meshnet = self.Meshnet(self)
        self.network = self.Network(self)

    def connect(self):
        self.client.connect(self.hostname, 22, username=self.username, password=self.password)

    def exec_command(self, command: str) -> str:
        _, stdout, stderr = self.client.exec_command(command, timeout=10)
        if stdout.channel.recv_exit_status() != 0:
            msg = f'{stdout.read().decode()} {stderr.read().decode()}'
            raise RuntimeError(msg)
        return stdout.read().decode()

    # Sends file in the provided path to the ssh peer
    # path and remote_path MUST be different, otherwise an empty file will be uploaded (fails to read local file for some reason)
    def send_file(self, path: str, remote_path: str):
        with self.client.open_sftp() as sftp:
            sftp.put(path, remote_path)

    # Downloads file to the provided path from the ssh peer
    def download_file(self, remote_path: str, path: str):
        with self.client.open_sftp() as sftp:
            sftp.get(remote_path, path)

    def disconnect(self):
        self.client.close()

    class Daemon:
        def __init__(self, ssh_class_instance):
            self.ssh_class_instance: Ssh = ssh_class_instance

        def is_running(self):
            try:
                self.ssh_class_instance.exec_command("nordvpn status")
            except RuntimeError:
                return False
            else:
                return True

    class IO:
        def __init__(self, ssh_class_instance):
            self.ssh_class_instance: Ssh = ssh_class_instance

        def get_file_hash(self, file_path: str) -> str:
            return self.ssh_class_instance.exec_command(f"{lib.FILE_HASH_UTILITY} {file_path}").split()[0]

        def create_directory(self, file_count: int, name_suffix: str = "", parent_dir: str = "", file_size: str = "1K") -> Directory:
            """
            Creates a temporary directory on remote peer and populates it with a specified number of files.

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
            dir_path = self.ssh_class_instance.exec_command(f"mktemp -d {f'{parent_dir}/tmp.XXXXXX' if parent_dir else ''}").split()[0]
            paths = []
            transfer_paths = []
            filenames = []
            filehashes = []

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

                self.ssh_class_instance.exec_command(f"fallocate -l {file_size} {dir_path}/{filename}")

                hash_output = self.ssh_class_instance.exec_command(f"sha256sum {path}").strip().split()[0]  # Only take the hash part of the output
                filehashes.append(hash_output)

            return Directory(dir_path, paths, transfer_paths, filenames, filehashes)

    class Network:
        def __init__(self, ssh_class_instance):
            self.ssh_class_instance: Ssh = ssh_class_instance

        def _is_internet_reachable(self, retry=5) -> bool:
            """Returns True when remote host is reachable by it's public IP."""
            i = 0
            while i < retry:
                try:
                    return "icmp_seq=" in self.ssh_class_instance.exec_command("ping -c 1 -w 1 1.1.1.1")
                except RuntimeError:
                    time.sleep(1)
                    i += 1
            return False

        def _is_dns_not_resolvable(self, retry=5) -> bool:
            """Returns True when domain resolution is not working."""
            for _ in range(retry):
                try:
                    with pytest.raises(RuntimeError) as ex:
                        self.ssh_class_instance.exec_command("ping -c 1 -w 1 nordvpn.com")

                    return "Network is unreachable" in str(ex) or \
                        "Name or service not known" in str(ex) or \
                        "Temporary failure in name resolution" in str(ex)
                except RuntimeError as ex:
                    time.sleep(1)
            return False

        def is_not_available(self, retry=5) -> bool:
            """Returns True when network access is not available."""
            return not self._is_internet_reachable(retry) and self._is_dns_not_resolvable(retry)

        def ping(self, target: str, retry=5) -> bool:
            i = 0
            while i < retry:
                try:
                    return "icmp_seq=" in self.ssh_class_instance.exec_command(f"ping -c 1 -w 1 {target}")
                except RuntimeError:
                    time.sleep(1)
                    i += 1
            return False

        def get_external_device_ip(self) -> str:
            """Returns external device IP."""
            cmd = f"wget -qO- {lib.API_EXTERNAL_IP}"
            output = self.ssh_class_instance.exec_command(cmd)

            try:
                json_data = json.loads(output)
                external_ip = json_data.get("ip", "")
                return external_ip
            except json.JSONDecodeError as e:
                print("Error decoding JSON:", e)

    class Meshnet:
        def __init__(self, ssh_class_instance):
            self.ssh_class_instance = ssh_class_instance

        def set_permissions(self, peer: str, routing: bool | None = None, local: bool | None = None, incoming: bool | None = None, fileshare: bool | None = None):
            def bool_to_permission(permission: bool) -> str:
                if permission:
                    return "allow"
                return "deny"

            # ignore any failures that might occur when permissions are already configured to the desired value
            if routing is not None:
                with contextlib.suppress(Exception):
                    self.ssh_class_instance.exec_command(f"nordvpn mesh peer routing {bool_to_permission(routing)} {peer}")

            if local is not None:
                with contextlib.suppress(Exception):
                    self.ssh_class_instance.exec_command(f"nordvpn mesh peer local {bool_to_permission(local)} {peer}")

            if incoming is not None:
                with contextlib.suppress(Exception):
                    self.ssh_class_instance.exec_command(f"nordvpn mesh peer incoming {bool_to_permission(incoming)} {peer}")

            if fileshare is not None:
                with contextlib.suppress(Exception):
                    self.ssh_class_instance.exec_command(f"nordvpn mesh peer fileshare {bool_to_permission(fileshare)} {peer}")

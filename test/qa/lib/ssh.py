import contextlib

import paramiko


class Ssh:
    def __init__(self, hostname: str, username: str, password: str):
        self.client = paramiko.SSHClient()
        self.hostname = hostname
        self.username = username
        self.password = password
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        self.meshnet = self.Meshnet(self)

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


    class Meshnet:
        def __init__(self, ssh_class_instance):
            self.ssh_class_instance = ssh_class_instance

        def set_permissions(self, peer: str, routing: bool = None, local: bool = None, incoming: bool = None, fileshare: bool = None):
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

import os
import json
import subprocess
import requests
from lib.remote_config import RemoteConfig

LOCAL_CACHE_DIR = "/var/lib/nordvpn/conf"
REMOTE_DIR = os.environ.get("REMOTE_DIR")
REMOTE_FILES = os.environ.get("REMOTE_FILES")


class RemoteConfigManager:
    """Manages remote, local, and backup config retrieval, hash checking, and caching."""

    def __init__(self, env: str, cache_dir: str, default_config: dict) -> None:
        self.env = env
        self.cache_dir = cache_dir or LOCAL_CACHE_DIR
        self.default_config = default_config

    def set_permissions_cache_dir(self) -> bool:
        """
        Sets recursive permissions of the "cache_dir" directory to "775".

        This method ensures that the "cache_dir" directory and all its contents
        are fully accessible (read, write, execute) by all users. If the directory does
        not exist, it logs a message and exits without performing any operation.

        :return: True if successful, False otherwise.
        """
        if not os.path.exists(self.cache_dir):
            print(f"Directory '{self.cache_dir}' does not exist.")
            return False
        subprocess.run(["sudo", "chmod", "-R", "775", self.cache_dir], check=True)
        print(f"Permissions set recursively to 775 for {self.cache_dir}.")
        return True

    def get_local_files(self):
        """
        Retrieves the full file paths of all files in the "cache_dir" directory.

        This method ensures that "cache_dir" permissions are correct by calling
        "set_permissions_cache_dir" before listing the files.

        :return: A list of full file paths for all files in the "cache_dir".
        """
        self.set_permissions_cache_dir()
        config_files = os.listdir(self.cache_dir)
        full_file_paths = [
            os.path.join(self.cache_dir, filename) for filename in config_files
        ]

        print(f"Local files: {config_files}")
        return full_file_paths

    def get_local_hash_files(self):
        """
        Retrieve full file paths of local configuration files matching "*-hash.json".

        This method scans the local configuration directory and identifies files whose
        names match the pattern `*-hash.json`. Only files in the top-level directory are
        considered, and their absolute paths are returned.

        :return: A list of full file paths for files named "*-hash.json".
        """
        mask = "-hash.json"
        hash_files = []
        for rc_file in self.get_local_files():
            if rc_file.lower().endswith(mask):
                hash_files.append(rc_file)

        print(f"Local hash files: {hash_files}")
        return hash_files

    def get_local_config_files(self):
        """
        Retrieve full file paths of local configuration files excluding "*-hash.json".

        This method scans the local configuration directory (`cache_dir`) and selects files
        that do not match the naming pattern "*-hash.json". Only files in the top-level
        directory are considered, and their full file paths are returned.

        :return: A list of full file paths for all config files except those named "*-hash.json".
        """
        mask = "-hash.json"
        config_files = []
        for rc_file in self.get_local_files():
            if not rc_file.lower().endswith(mask):
                config_files.append(rc_file)

        print(f"Local config files: {config_files}")
        return config_files

    def read_config(self, file_path: str) -> RemoteConfig:
        """
        Reads a locally cached config file and returns it as a "RemoteConfig" instance.

        This method reads the provided JSON file from the given file path, deserializes the
        JSON content, and wraps it into a "RemoteConfig" object.

        :param file_path: The path to the JSON file to read.

        :return: A RemoteConfig instance.
        """
        with open(file_path) as f:
            data = json.load(f)
        return RemoteConfig(data)

    def get_remote_configs(self) -> dict:
        """
        Downloads remote config JSONs by concatenating REMOTE_DIR with each file in REMOTE_FILES.

        For each file in REMOTE_FILES, a request is made to construct the full URL
        (REMOTE_DIR + REMOTE_FILE). The function then fetches the corresponding JSON
        data for each URL and returns a dictionary where the key is the full URL and
        the value is the parsed RemoteConfig instance.

        :raises ValueError: If REMOTE_DIR is not set.
        :raises requests.RequestException: For network-related issues.
        :raises json.JSONDecodeError: If the response is not valid JSON.

        :return: A dict mapping remote URLs to RemoteConfig instances.
        """

        if not REMOTE_DIR:
            raise ValueError("REMOTE_DIR environment variable is not set.")

        if not REMOTE_FILES:
            raise ValueError("REMOTE_FILES environment variable is not set.")

        remote_files = REMOTE_FILES.split(",")
        configs = {}

        for rc_file in remote_files:
            full_url = os.path.join(REMOTE_DIR, rc_file.strip())
            try:
                print(f"Fetching config from: {full_url}")
                response = requests.get(full_url, timeout=5)
                response.raise_for_status()
                data = response.json()
                configs[full_url] = RemoteConfig(data)

            except requests.RequestException as e:
                print(f"Failed to fetch {full_url}: {e}")
                raise RuntimeError(e) from e
            except requests.JSONDecodeError as e:
                print(f"Invalid JSON response from {full_url}: {e}")
                raise RuntimeError(e) from e

        return configs

    def delete_file(self, file_path: str) -> None:
        """
        Deletes file using 'sudo rm'.

        :param file_path: The path to the file to delete.

        :raises subprocess.CalledProcessError: If file deletion fails.
        """
        print(f"Deleting file: {file_path}")
        try:
            subprocess.run(["sudo", "rm", file_path], check=True)
        except subprocess.CalledProcessError as e:
            print(f"Got an error while deleting {file_path}: {e}")

import copy
import json
import subprocess
from typing import Any


class RemoteConfig:
    """Wrapper for config data, supporting safe access, mutation, and history tracking."""

    def __init__(self, data: dict) -> None:
        self._data = copy.deepcopy(data)
        self._history: list[dict] = [copy.deepcopy(data)]

    def get(self, key: str, default: str| None = None) -> Any:
        """
        Retrieve the value associated with a given key, or return a default value if the key is not found.

        :param key: The key to retrieve.
        :param default: The default value to return if the key is not found.

        :return: The value associated with the key, or the default value if the key is not found.
        """
        return self._data.get(key, default)

    def set(self, key: str, value: Any) -> None:
        """
        Update the value associated with a key in the config and store the current state in the history.

        :param key: The key to update.
        :param value: The value to store.
        """
        self._data[key] = value
        self._history.append(copy.deepcopy(self._data))

    def delete(self, key: str) -> None:
        """
        Remove a key from the config if it exists and store the current state in the history.

        :param key: The key to remove.
        """
        if key in self._data:
            del self._data[key]
            self._history.append(copy.deepcopy(self._data))

    def as_dict(self) -> list[dict]:
        """
        Return a deep copy of the current config data as a dictionary.

        :return: A deep copy of the current config data.
        """
        return copy.deepcopy(self._data)

    def version(self) -> int:
        """
        Retrieve the version number of the config.

        :return: The version number.
        """
        return self._data.get("version")

    def history(self) -> list[dict]:
        """
        Retrieve the full history of config changes as a list of dictionaries.

        :return: The full history of config changes as a list of dictionaries.
        """
        return copy.deepcopy(self._history)

    def save_to_file(self, file_path: str) -> None:
        """
        Save the current config to a file as JSON using "subprocess.run".

        :param file_path: The path to the file to save.

        :raises subprocess.CalledProcessError: If the "subprocess.run" fails to execute properly.
        """
        json_data = json.dumps(self._data, indent=4)

        try:
            subprocess.run(
                ["sudo", "tee", file_path],
                input=json_data.encode("utf-8"),
                check=True,
            )
            print(f"config saved successfully to {file_path}")
        except subprocess.CalledProcessError as e:
            print(f"Failed to save config to {file_path}: {e}")
            raise

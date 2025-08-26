import hashlib
import uuid

import pytest

from lib import daemon
from lib.log_reader import LogReader

RC_REMOTE_MESSAGES = [
    "[Info] feature [meshnet] remote config downloaded to: /var/lib/nordvpn/conf",
    "[Info] feature [libtelio] remote config downloaded to: /var/lib/nordvpn/conf",
    "[Info] feature [nordvpn] remote config downloaded to: /var/lib/nordvpn/conf",
]

RC_LOCAL_MESSAGES = [
    "[Info] feature [meshnet] config loaded from: /var/lib/nordvpn/conf",
    "[Info] feature [libtelio] config loaded from: /var/lib/nordvpn/conf",
    "[Info] feature [nordvpn] config loaded from: /var/lib/nordvpn/conf",
]

RC_INITIAL_RUN_MESSAGES = RC_REMOTE_MESSAGES + RC_LOCAL_MESSAGES


@pytest.fixture
def initialized_app_with_remote_config(
    clean_cache_files, # noqa: ARG001
    daemon_log_cursor: int,
    nordvpnd_scope_function, # noqa: ARG001
    daemon_log_reader: LogReader,
) -> None:
    """
    Fixture to verify that the application consumes the remote config files and logs the relevant messages.

    This fixture ensures the application starts correctly with remote config
    files successfully downloaded and verifies that the required log messages have
    been recorded in the daemon log file. It raises an `AssertionError` if the
    expected log messages are not found, indicating that the configuration might
    not have been applied correctly.

    :raises AssertionError: If the expected log messages are not found in the daemon log.
    """
    assert daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES, cursor=daemon_log_cursor
    ), "Expected 'RC_INITIAL_RUN_MESSAGES' to appear in the daemon logs."


def test_remote_config_consumed_and_logged(initialized_app_with_remote_config) -> None:
    """
    Verify that the remote config is consumed and relevant log messages are recorded.

    Test steps:
    1. Verify that remote config is consumed by checking for specific log messages.
    2. Use the provided cursor to ensure the messages appear as expected in the daemon.log.

    :raises AssertionError: If the expected log messages are not found in the daemon log.
    """
    pass


def test_local_files_removal_and_daemon_restart(
    initialized_app_with_remote_config, # noqa: ARG001
    daemon_log_reader,
    rc_config_manager,
) -> None:
    """
    Test the application's behavior when all files in `LOCAL_CACHE_DIR` are deleted and the daemon is restarted.

    This test verifies that the application handles the absence of locally cached files correctly
    by checking if the appropriate log messages are produced after a restart.

    Test steps:
    1. Confirm that the remote config is present by checking relevant log messages.
    2. Record the current position in the daemon.log using the cursor.
    3. Change permissions recursively on the local config folder to allow deletion.
    4. Stop the daemon service to prepare for config deletion.
    5. Delete the local config folder.
    6. Start the daemon service.
    7. Verify that the application logs the appropriate messages
       ("RC_INITIAL_RUN_MESSAGES") after the daemon startup.

    :raises AssertionError: If the expected log messages are not found after deleting the folder.
    """

    rc_files = rc_config_manager.get_local_files()
    cursor = daemon_log_reader.get_cursor()

    daemon.stop()

    for rc_file in rc_files:
        rc_config_manager.delete_file(rc_file)

    daemon.start()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES, cursor=cursor
    )


def test_local_hash_files_removal_and_daemon_restart(
    initialized_app_with_remote_config, # noqa: ARG001
    daemon_log_reader,
    rc_config_manager,
) -> None:
    """
    Test to verify the application's behavior when local hash config files are deleted and the daemon is restarted.

    This test ensures the application performs as expected when critical hash configuration files
    are removed. It validates that the daemon's behavior remains consistent and checks for the
    presence of specific log messages after the daemon restarts.

    Test Steps:
    1. Retrieve all local config files using the "rc_config_manager".
    2. Identify and remove files that match the naming pattern "*-hash.json".
    3. Stop the daemon after identifying and deleting the relevant files.
    4. Restart the daemon after modifications to the local config files.
    5. Verify that the application logs the appropriate messages
       ("RC_INITIAL_RUN_MESSAGES") after the daemon startup.

    :raises AssertionError: If the expected log messages are not found after deleting hash files.
    """

    rc_hash_files = rc_config_manager.get_local_hash_files()

    cursor = daemon_log_reader.get_cursor()

    daemon.stop()

    for hash_file in rc_hash_files:
        rc_config_manager.delete_file(hash_file)

    daemon.start()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES, cursor=cursor
    )


def test_disable_remote_config_download_and_verify_log_messages_not_found(
    initialized_app_with_remote_config, # noqa: ARG001
    disable_remote_endpoint, # noqa: ARG001
    daemon_log_reader: LogReader,
) -> None:
    """
    Test to verify that the application does not fetch remote configuration files from a disabled remote endpoint.

    This test ensures that the application respects the connection block to the specified
    remote endpoint by validating that log messages related to downloading configurations
    are absent. It further ensures that the missing messages align with the expected results.

    Test steps:
        1. Disable connections to "remote endpoint" using the provided
           "disable_remote_endpoint" fixture. This ensures the app cannot fetch remote configs.
        2. Read the daemon.log starting from the given cursor position.
        3. Wait for the expected success messages ("RC_INITIAL_RUN_MESSAGES") not to appear in the log file.
        4. Verify that the expected success messages ("RC_INITIAL_RUN_MESSAGES") are the same as not found ones.

    :raises AssertionError: If log messages related to downloading configs are found.
    :raises AssertionError: If the content of missing messages does not match expectations.
    """

    daemon.restart()

    cursor = daemon_log_reader.get_cursor()

    res, not_found = daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES,
        cursor=cursor,
        return_not_found=True,
    )

    assert not res, "Expected success messages to not be found in the logs."
    assert (
        not_found == RC_INITIAL_RUN_MESSAGES
    ), "The missing messages do not match the expected RC_INITIAL_RUN_MESSAGES."


@pytest.mark.parametrize(
    "hash_value",
    [
        "",
        hashlib.sha256(uuid.uuid4().bytes).hexdigest(),
        None,
        "123",
        uuid.uuid4().hex,
    ],
    ids=[
        "empty_string",
        "random_sha256_hash",
        "null_value",
        "short_string",
        "random_uuid_hex",
    ],
)
def test_hash_modification_and_daemon_restart(
    initialized_app_with_remote_config, # noqa: ARG001
    daemon_log_reader,
    rc_config_manager,
    hash_value,
):
    """
    Test the application's behavior and daemon.log output after modifying hash values in local config files.

    This test validates the application's ability to handle changes to hash values
    stored in local configuration files (`*-hash.json`). It ensures that updates to
    these files are correctly applied, persisted, and reflected in the daemon logs
    after a restart.

    Test Steps:
    1. Retrieve all local config files using the "rc_config_manager".
    2. Identify config files that match the naming pattern "*-hash.json".
    3. For each matching file:
        - Record the current "hash" value.
        - Update the "hash" value with a different hash value.
        - Save the modified config back to the file.
        - Verify that the saved config reflects the updated "hash".
    4. Stop the daemon, and restart it after modifications.
    5. Verify that modifying the hash values triggers expected daemon log messages
       ("RC_INITIAL_RUN_MESSAGES") after restarting the daemon.

    :raises AssertionError: If hash values are not modified or daemon logs do not
        contain the expected messages.
    """
    rc_hash_files = rc_config_manager.get_local_hash_files()
    rc_hash_data_before, rc_hash_data_after = [], []

    cursor = daemon_log_reader.get_cursor()

    daemon.stop()

    for hash_file in rc_hash_files:
        config = rc_config_manager.read_config(hash_file)
        rc_hash_data_before.append((hash_file, config.get("hash")))

        # Set new hash value
        config.set("hash", hash_value)
        config.save_to_file(hash_file)

        hash_value_after = rc_config_manager.read_config(hash_file).get("hash")
        rc_hash_data_after.append((hash_file, hash_value_after))

    assert rc_hash_data_before != rc_hash_data_after, "Expected different hash values."

    daemon.start()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES, cursor=cursor
    ), "Expected 'RC_INITIAL_RUN_MESSAGES' to appear in the daemon logs."

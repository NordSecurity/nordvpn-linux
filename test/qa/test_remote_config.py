import hashlib
import json
import subprocess
import time
import uuid

import pytest

from lib import daemon
from lib.log_reader import LogReader

from lib.remote_config_manager import LOCAL_CACHE_DIR

RC_REMOTE_MESSAGES = "[Info] feature [{}] remote config downloaded to: /var/lib/nordvpn/conf"
RC_LOCAL_MESSAGES = "[Info] feature [{}] config loaded from: /var/lib/nordvpn/conf"
RC_REQUEST_MESSAGES = "Request:  GET https://downloads.nordcdn.com/apps/linux/config/dev/{}-hash.json"
SERVICES_TO_BE_CHECK = ["nordvpn", "libtelio", "meshnet"]
RC_INITIAL_RUN_MESSAGES = [RC_REMOTE_MESSAGES.format(service) for service in SERVICES_TO_BE_CHECK] + [
    RC_LOCAL_MESSAGES.format(service) for service in SERVICES_TO_BE_CHECK
]
RC_MESHNET_CONFIG_FILE = "meshnet.json"
RC_MESHNET_HASH_FILE = "meshnet-hash.json"


@pytest.fixture
def initialized_app_with_remote_config(
    clean_cache_files,  # noqa: ARG001
    daemon_log_cursor: int,
    nordvpnd_scope_function,  # noqa: ARG001
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
    initialized_app_with_remote_config,  # noqa: ARG001
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

    assert daemon_log_reader.wait_for_messages(messages=RC_INITIAL_RUN_MESSAGES, cursor=cursor)


def test_local_hash_files_removal_and_daemon_restart(
    initialized_app_with_remote_config,  # noqa: ARG001
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

    assert daemon_log_reader.wait_for_messages(messages=RC_INITIAL_RUN_MESSAGES, cursor=cursor)


def test_disable_remote_config_download_and_verify_log_messages_not_found(
    initialized_app_with_remote_config,  # noqa: ARG001
    disable_remote_endpoint,  # noqa: ARG001
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
    initialized_app_with_remote_config,  # noqa: ARG001
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


@pytest.mark.skip(reason="Not implemented mock CDN")
@pytest.mark.parametrize(
    "tcid, error_message",
    [
        pytest.param("LVPN-8452", "error: downloading main hash file:", id="no_cache"),
        pytest.param(
            "LVPN-8453", "failed downloading feature [ nordvpn ] remote config: downloading main file", id="no_config"
        ),
    ],
)
def test_remote_config_cdn_unavailable_(
    tcid,
    error_message,
    initialized_app_with_remote_config,  # noqa: ARG001
    disable_remote_endpoint,  # noqa: ARG001
    set_custom_timeout_for_rc_retry_scheme,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    stop_nordvpnd,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :tcid       {tcid}

    :details    Verify that app handles the case, when remote files are not available in CDN and no cache files are locally
    :keywords   RC

    :preconditions
        - # remote file is removed from CDN according to use case
        - # No remote config files are cached on local disk
        - # nordvpnd is stopped
        - # Timeout in daemon service is set to 1 min

    :steps
        - # Start nordvpnd daemon
        - # Check journalctl for error logs related to remote config
        - # Wait for next check
        - # Check journalctl for error logs related to remote config

    :endsteps
        - # Restore original config in daemon config file

    :expected
        - # Nordvpn daemon shows error related to remote config
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows error related to remote config
    """
    first_time_mark = daemon_log_reader.get_cursor()
    daemon.restart()

    assert daemon_log_reader.wait_for_messages(
        RC_REMOTE_MESSAGES.format(error_message), cursor=first_time_mark, timeout=90
    ), f"Couldn't found error logs"

    second_time_mark = daemon_log_reader.get_cursor()

    assert daemon_log_reader.wait_for_messages(
        RC_REMOTE_MESSAGES.format(error_message), cursor=second_time_mark, timeout=90
    ), f"Couldn't found error logs"


def test_remote_config_download_config_on_start(
    initialized_app_with_remote_config,  # noqa: ARG001
    set_custom_timeout_for_rc_retry_scheme,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    stop_nordvpnd,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :tcid       LVPN-8456

    :details    Verify that app verifies local config files and download then
    :keywords   RC

    :preconditions
        - # Config files are available on CDN
        - # No remote config files are cached on local disk
        - # nordvpnd is stopped
        - # Timeout in daemon service is set to 1 min

    :steps
        - # Start nordvpnd daemon
        - # Check journalctl for download logs related to remote config
        - # Wait for next check
        - # Check journalctl for request logs related to remote config

    :endsteps
        - # Restore original config in daemon config file

    :expected
        - # Nordvpn daemon shows download related to remote config
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows request logs related to remote config
    """
    missed_services_config = []
    daemon.start()

    for service_config in SERVICES_TO_BE_CHECK:
        if not daemon_log_reader.wait_for_messages(RC_REMOTE_MESSAGES.format(service_config), timeout=90):
            missed_services_config.append(service_config)

    assert not missed_services_config, f"Couldn't found download logs related to {missed_services_config}"

    expected_service_config = SERVICES_TO_BE_CHECK
    time_mark = time.time()

    while time.time() < time_mark + 90:
        second_time_mark = daemon_log_reader.get_cursor()
        for service_config in SERVICES_TO_BE_CHECK:
            for line in daemon_log_reader.get_partial_log(cursor=second_time_mark):
                if "Request" in line and "GET" in line and service_config + "-hash.json" in line:
                    expected_service_config.remove(line)
        time.sleep(10)

    assert not expected_service_config, f"Couldn't found request logs related to {expected_service_config}"


@pytest.mark.parametrize(
    "tcid, is_config_different",
    [
        pytest.param("LVPN-8477", False, id="equal_remote_config"),
        pytest.param(
            "LVPN-8500", True, id="different_remote_config", marks=pytest.mark.skip(reason="Not implemented mock CDN")
        ),
    ],
)
def test_remote_config_attempts_config_(
    tcid,  # noqa: ARG001
    is_config_different,  # noqa: ARG001
    initialized_app_with_remote_config,  # noqa: ARG001
    set_custom_timeout_for_rc_retry_scheme,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    stop_nordvpnd,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :tcid       {tcid}

    :details    Verify that app verifies local config files and download then with additional verification of remote config
    :keywords   RC

    :preconditions
        - # Config files are available on CDN
        - # No remote config files are cached on local disk
        - # nordvpnd is stopped
        - # Timeout in daemon service is set to 1 min

    :steps
        - # Start nordvpnd daemon
        - # Check journalctl for download logs related to remote config
        - # Check date of config files
        - # Wait for next check
        - # Check journalctl for download logs related to remote config
        - # Check date of config files

    :endsteps
        - # Restore original config in daemon config file

    :expected
        - # Nordvpn daemon shows download related to remote config
        - # Files are downloaded
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows download logs related to remote config
        - # Files are downloaded/not downloaded according to remote config files
    """
    command_for_conf_file = f'sudo find {LOCAL_CACHE_DIR} -type f -exec stat -c "%a %y %n" {{}} \\;'
    chmod_to_be_found = "600"
    missed_services_config = []

    daemon.restart()

    for service_config in SERVICES_TO_BE_CHECK:
        if not daemon_log_reader.wait_for_messages(RC_REMOTE_MESSAGES.format(service_config), timeout=90):
            missed_services_config.append(service_config)

    assert not missed_services_config, f"Couldn't found download logs related to {missed_services_config}"

    conf_files_data = subprocess.run(command_for_conf_file, shell=True, check=True, capture_output=True, text=True)

    wrong_permissions_files = []
    for line in conf_files_data.stdout.splitlines():
        if chmod_to_be_found not in line:
            wrong_permissions_files.append(line)

    assert (
        not missed_services_config
    ), f"Found files that do not have {chmod_to_be_found} chmod: {wrong_permissions_files}"

    expected_service_config = SERVICES_TO_BE_CHECK
    time_mark = time.time()

    while time.time() < time_mark + 90:
        second_time_mark = daemon_log_reader.get_cursor()
        for service_config in SERVICES_TO_BE_CHECK:
            for line in daemon_log_reader.get_partial_log(cursor=second_time_mark):
                if "Request" in line and "GET" in line and service_config + "-hash.json" in line:
                    expected_service_config.remove(line)
        time.sleep(10)

    assert expected_service_config, f"Couldn't found download logs related to {expected_service_config}"

    conf_files_data_after_attempt = subprocess.run(
        command_for_conf_file, shell=True, check=True, capture_output=True, text=True
    )
    different_files = []

    for line in conf_files_data_after_attempt.stdout.splitlines():
        if is_config_different:
            for nordvpn_file in ("nordvpn.json", "nordvpn-hash.json"):
                if nordvpn_file not in (
                    list(
                        set(conf_files_data_after_attempt.stdout.splitlines())
                        - set(conf_files_data.stdout.splitlines())
                    )
                ):
                    different_files.append(line)
            break
        else:
            if line not in conf_files_data.stdout.splitlines():
                different_files.append(line)

    assert not different_files, f"Found some files mismatch: {different_files}"


@pytest.mark.parametrize(
    "tcid, parameter, value, additional_log_verification",
    [
        pytest.param("LVPN-8544", "value", False, False, id="set_value_false"),
        pytest.param("LVPN-8522", "rollout", None, False, id="remove_rollout"),
        pytest.param("LVPN-8782", "app_version", "*", False, id="set_app_version_any"),
        pytest.param("LVPN-8518", "app_version", "", False, id="set_app_version_empty"),
        pytest.param(
            "LVPN-8519",
            "app_version",
            "y",
            "invalid version constraint: improper constraint: y",
            id="set_app_version_invalid_version",
        ),
    ],
)
def test_remote_config_change_local_meshnet_config_settings_(
    tcid,  # noqa: ARG001
    parameter,
    value,
    additional_log_verification,
    backup_restore_rc_config_files,
    rc_config_manager,
    daemon_log_reader,  # noqa: ARG001
    set_custom_timeout_for_rc_retry_scheme,  # noqa: ARG001
    initialized_app_with_remote_config,  # noqa: ARG001
):
    """
    :tcid       {tcid}

    :details    Verify that app change behavior according to config

    :preconditions
        - # No remote config files are cached on local disk
        - # nordvpnd is stopped
        - # Timeout in daemon service is set to 1 min

    :steps
        - # Change in meshnet.json file "value": false
        - # Start nordvpnd daemon
        - # Check journalctl for loading config logs
        - # Check that command "nordvpn meshnet" returns "Command 'meshnet' doesn't exist."

    :endsteps
        - # Restore original config in daemon config file

    :expected
        - # Value is set to False
        - # Config file was loaded by app
        - # Meshnet option is unavailable
    """
    if not rc_config_manager.set_permissions_cache_dir():
        pytest.skip("Directory doesn't exist")

    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}", "r") as file:
        config = json.load(file)

    if value is not None:
        config["configs"][0]["settings"][0][parameter] = value
    elif value is None:
        del config["configs"][0]["settings"][0][parameter]

    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}", "w") as file:
        json.dump(config, file, indent=4)

    sha_sun_hash = subprocess.run(
        f"sha256sum {LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}", capture_output=True, shell=True, text=True
    ).stdout.split()[0]

    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_HASH_FILE}", "w") as file:
        json.dump({"hash": sha_sun_hash}, file)

    cursor = daemon_log_reader.get_cursor()
    daemon.restart()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_LOCAL_MESSAGES.format("meshnet"), cursor=cursor, timeout=90
    ), "Couldn't found log of loading modified json file"

    if additional_log_verification:
        assert daemon_log_reader.wait_for_messages(
            messages=additional_log_verification, cursor=cursor, timeout=90
        ), f"Couldn't found log {additional_log_verification}"

    if value is False:
        with pytest.raises(subprocess.CalledProcessError):
            subprocess.run("nordvpn meshnet", shell=True, text=True, check=True)
    elif value is None:
        try:
            subprocess.run("nordvpn meshnet", shell=True, text=True, check=True)
        except subprocess.CalledProcessError:
            pytest.fail(reason="nordvpn meshnet is not enabled")

import hashlib
import json
import subprocess
import time
import uuid
import os
from typing import Any

import pytest
import sh

from lib import daemon, set_technology_and_protocol, STANDARD_TECHNOLOGIES_NO_NORDWHISPER, Defer
from lib.log_reader import LogReader
from lib.remote_config_manager import LOCAL_CACHE_DIR, RemoteConfigManager
from lib.daemon import (
    Env,
    disable_rc_local_config_usage,
    enable_rc_local_config_usage,
    remove_rc_retry_custom_time,
    set_rc_retry_custom_time,
)
from constants import RC_TIMEOUT
from test_killswitch import MSG_KILLSWITCH_ON

RC_INVALID_VERSION_MESSAGE = "invalid version constraint: improper constraint: y"
RC_REMOTE_MESSAGES = [
    f"[Info] feature [meshnet] remote config downloaded to: {LOCAL_CACHE_DIR}",
    f"[Info] feature [libtelio] remote config downloaded to: {LOCAL_CACHE_DIR}",
    f"[Info] feature [nordvpn] remote config downloaded to: {LOCAL_CACHE_DIR}",
    f"[Info] feature [nordwhisper] remote config downloaded to: {LOCAL_CACHE_DIR}",
]

RC_LOCAL_MESSAGES = [
    f"[Info] feature [meshnet] config loaded from: {LOCAL_CACHE_DIR}",
    f"[Info] feature [libtelio] config loaded from: {LOCAL_CACHE_DIR}",
    f"[Info] feature [nordvpn] config loaded from: {LOCAL_CACHE_DIR}",
    f"[Info] feature [nordwhisper] config loaded from: {LOCAL_CACHE_DIR}",
]

RC_USE_ECH_MESSAGE = ["TLS is using ECH"]

RC_REQUEST_MESSAGES = "Request:  GET https://downloads.nordcdn.com/apps/linux/config/dev/{}-hash.json"
FEATURE_TO_BE_CHECK = ["nordvpn", "libtelio", "meshnet"]
RC_INITIAL_RUN_MESSAGES = RC_REMOTE_MESSAGES + RC_LOCAL_MESSAGES
RC_MESHNET_CONFIG_FILE = "meshnet.json"
RC_MESHNET_HASH_FILE = "meshnet-hash.json"
RC_NORDWHISPER_CONFIG_FILE = "nordwhisper.json"
RC_NORDWHISPER_HASH_FILE = "nordwhisper-hash.json"

RC_USE_LOCAL_CONFIG_MESSAGE = ["[Info] Ignoring remote config, using only local"]


def set_ech_for_local_config(enabled: bool = True,
                             entries_to_remove: list[tuple[str, Any]] | None = None,) -> None:
    """
    Enable or disable ECH in the local cached RC nordwhisper file.

    :param enabled: enable/disable ECH in the local cached RC nordwhisper file.
    :param entries_to_remove: JSON fields to remove(needed for testing purpose) from configs section
    """

    if entries_to_remove is None:
        entries_to_remove = []

    config_path = f"{LOCAL_CACHE_DIR}/{RC_NORDWHISPER_CONFIG_FILE}"
    hash_path = f"{LOCAL_CACHE_DIR}/{RC_NORDWHISPER_HASH_FILE}"

    with open(config_path, encoding="utf-8") as config_file:
        config_data = json.load(config_file)

    for config_entry in config_data.get("configs", []):
        if config_entry.get("name") == "enable_ech":
            for setting in config_entry.get("settings", []):
                setting["value"] = enabled
                setting["app_version"] = ">=4.4.0"

    for config_entry in config_data.get("configs", []):
        for key, expected_value in entries_to_remove:
            if config_entry.get(key) == expected_value:
                config_entry.pop(key, None)

    with open(config_path, "w", encoding="utf-8") as config_file:
        json.dump(config_data, config_file, indent=4)

    sha256 = sh.sha256sum(config_path).stdout.decode("utf-8").split()[0]
    with open(hash_path, "w", encoding="utf-8") as hash_file:
        json.dump({"hash": sha256}, hash_file)


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
    been recorded in the daemon log file. It raises an 'AssertionError' if the
    expected log messages are not found, indicating that the configuration might
    not have been applied correctly.

    :raises AssertionError: If the expected log messages are not found in the daemon log.
    """
    assert daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES, cursor=daemon_log_cursor
    ), f"Expected all {RC_INITIAL_RUN_MESSAGES} to appear in the daemon log."


@pytest.fixture
def enable_local_config_in_service():
    """
    Pytest fixture to temporarily inject 'RC_USE_LOCAL_CONFIG=1' environment variable into the 'nordvpnd' service file.

    This fixture injects 'export RC_USE_LOCAL_CONFIG=1' into the 'nordvpn' service file, simulating enabling the local config usage for tests.
    After the test, it removes the injected line and reloads the systemd daemon to restore the initial system state.
    """
    enable_rc_local_config_usage()

    yield

    disable_rc_local_config_usage()


@pytest.fixture
def set_custom_rc_retry_time_in_service(daemon_log_reader: LogReader):
    """
    Pytest fixture to temporarily inject 'RC_LOAD_TIME_MIN=1' environment variable into the 'nordvpnd' service file.

    After the test, it removes the injected line and reloads the systemd daemon to restore the initial system state.
    """
    cursor = set_rc_retry_custom_time(daemon_log_reader)
    if not daemon_log_reader.wait_for_messages(
        [f"[Info] remote config download job time period: {RC_TIMEOUT}m0s"], cursor=cursor
    ):
        pytest.fail("Service doesn't applied new time period.")

    yield

    remove_rc_retry_custom_time()

def check_log_for_request_get_messages(daemon_log_reader: LogReader) -> None:
    """Function to verify that nordvpn daemon attempts to do get request for hash files"""
    expected_service_config = FEATURE_TO_BE_CHECK.copy()
    cursor = daemon_log_reader.get_cursor()
    time_mark = time.time()

    while time.time() < time_mark + 90:
        log_content = daemon_log_reader.get_partial_log(cursor=cursor)
        cursor = daemon_log_reader.get_cursor()

        if not log_content:
            time.sleep(2)
            continue
        for service_config in FEATURE_TO_BE_CHECK:
            for line in log_content.splitlines():
                if "Request" in line and "GET" in line and f"{service_config}-hash.json" in line:
                    try:
                        print(f"Found {service_config}-hash.json in line")
                        expected_service_config.remove(service_config)
                        break
                    except ValueError:
                        break
        if not expected_service_config:
            print("Found all hash files in lines")
            break
        time.sleep(2)

    assert not expected_service_config, f"Couldn't found request logs related to {expected_service_config}"


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
    Test the application's behavior when all files in 'LOCAL_CACHE_DIR' are deleted and the daemon is restarted.

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


def test_restart_causes_daemon_to_consume_local_files(
    initialized_app_with_remote_config,  # noqa: ARG001
    daemon_log_reader,
) -> None:
    """
    Test that after a daemon restart, remote configuration files are consumed from the local cache.

    This test verifies that when the daemon is restarted, it attempts to use locally cached
    remote configuration files.

    Test steps:
    1. Verifies application initial run.
    2. Record the current cursor position in the daemon.log.
    3. Restart the daemon.
    4. Wait for RC_INITIAL_RUN_MESSAGES in the daemon logs after restart.
    5. Verify that RC_REMOTE_MESSAGES messages are not found,
    confirming the remote files were used from local.

    :raises AssertionError: If any of the RC_INITIAL_RUN_MESSAGES are found in the daemon.log after daemon restart or
    if the missing log messages do not match RC_REMOTE_MESSAGES.
    """

    cursor = daemon_log_reader.get_cursor()

    daemon.restart()

    res, not_found = daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES,
        cursor=cursor,
        return_not_found=True,
    )

    assert not res, "Some of the RC_INITIAL_RUN_MESSAGES are found in the daemon.log after daemon restart"

    assert set(not_found) == set(
        RC_REMOTE_MESSAGES
    ), "The missing messages do not match the expected RC_REMOTE_MESSAGES."


def test_disable_remote_config_download_and_verify_log_messages_not_found(
    initialized_app_with_remote_config,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
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
    stored in local configuration files ('*-hash.json'). It ensures that updates to
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


def test_equivalence_of_local_and_remote_config_files(
    initialized_app_with_remote_config, rc_config_manager, env  # noqa: ARG001
):
    """
    Test that local config files and their corresponding remote config files are equivalent.

    This test retrieves all remote and local configu files for the given environment.
    For each local config file, it finds the corresponding remote config file by matching
    filenames.

    Test steps:
    1. Verifies application initial run.
    2. Retrieve remote and local config files for the specified environment.
    3. For each local config, find the remote config with the same filename.
    4. Compare their dictionary representations for equality.

     :raises AssertionError: If any local config does not match the corresponding remote config, or if mapping is incomplete.
    """
    remote_configs = rc_config_manager.get_remote_config_files(env=env)
    local_configs = rc_config_manager.get_local_config_files()

    mapping = {}
    for local_path in local_configs:
        filename = local_path.split("/")[-1]
        for remote_url in remote_configs:
            if remote_url.endswith(filename):
                mapping[local_path] = remote_url
                break

    assert len(local_configs) == len(mapping), (
        "Not all local config files matched a remote config: " f"{len(local_configs)} local, {len(mapping)} matched"
    )

    for local_path, remote_url in mapping.items():
        print(f"Comparing local: {local_path} to remote: {remote_url}")
        assert (
            rc_config_manager.read_config(local_path).as_dict() == remote_configs[remote_url].as_dict()
        ), f"Mismatch between {local_path} and {remote_url}"


def test_meshnet_feature_availability_based_on_remote_config(
    initialized_app_with_remote_config, rc_config_manager, env  # noqa: ARG001
):
    """
    Test that meshnet related CLI commands are unavailable when meshnet is disabled in the remote config and vice versa.

    This test checks the meshnet feature flag in the remote config file and verifies that all meshnet related
    CLI commands return the expected error message if the feature is disabled and do not return it when enabled.

    Test steps:
    1. Verifies application initial run.
    2. Retrieve the remote meshnet config for the current environment.
    3. Check if the meshnet feature is enabled using the remote configuration.
    4. For a list of meshnet related CLI commands, run each command and capture the output.
    5. Assert that the output contains (if disabled) or does not contain (if enabled) the expected error message.

    :raises AssertionError: If the meshnet config cannot be found in the remote configs.
    :raises AssertionError: If the CLI output does not match the expected result based on the meshnet flag.
    """

    remote_configs = rc_config_manager.get_remote_config_files(env=env)
    meshnet_config = None

    for url, config_obj in remote_configs.items():
        if url.endswith("meshnet.json"):
            meshnet_config = config_obj
            break

    assert meshnet_config is not None, "Meshnet config was not found in the remote configs."

    is_meshnet_enabled = meshnet_config.as_dict()["configs"][0]["settings"][0]["value"]

    print(f"is_meshnet_enabled: {is_meshnet_enabled}")

    command_and_expected_message = [
        (["nordvpn", "set", "meshnet", "on"], "Command 'meshnet' doesn't exist"),
        (["nordvpn", "meshnet", "peer", "list"], "Command 'meshnet' doesn't exist"),
        (["nordvpn", "fileshare", "list"], "Command 'fileshare' doesn't exist"),
        (["nordvpn", "set", "meshnet", "off"], "Command 'meshnet' doesn't exist"),
    ]

    for command, expected_error in command_and_expected_message:
        result = subprocess.run(command, stdout=subprocess.PIPE, text=True)
        output = result.stdout
        print(f"Command '{' '.join(command)}' returned: \n{output}")
        if is_meshnet_enabled:
            assert (
                expected_error not in output
            ), f"Command {command} output is not as expected for enabled meshnet. Output:\n{output}"
        else:
            assert (
                expected_error in output
            ), f"Command {command} output is not as expected for disabled meshnet. Output:\n{output}"


@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
def test_local_config_usage_via_systemd_env(
    initialized_app_with_remote_config,  # noqa: ARG001
    daemon_log_reader,
    enable_local_config_in_service,  # noqa: ARG001
):
    """
    Test that the environment variable 'RC_USE_LOCAL_CONFIG=1' into the 'nordvpnd' service causes the application to use only local config files.

    Steps:
      1. Verifies application initial run.
      2. Inject the environment variable 'RC_USE_LOCAL_CONFIG=1' into the 'nordvpnd' service.
      3. Reload the systemd daemon and restart the 'nordvpnd' to apply the change.
      4. Record the current daemon log cursor to mark the log position before the restart.
      5. Check the logs for the messages confirming local config usage.

    :raises AssertionError: If a message indicating local config usage is not found in the logs,
        if any remote initialization messages appear after enabling the local config environment variable,
        or if the set of missing messages does not match RC_REMOTE_MESSAGES.
    """

    cursor = daemon_log_reader.get_cursor()

    daemon.restart()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_USE_LOCAL_CONFIG_MESSAGE,
        cursor=cursor,
    )

    found_messages, missing_messages = daemon_log_reader.wait_for_messages(
        messages=RC_INITIAL_RUN_MESSAGES,
        cursor=cursor,
        return_not_found=True,
    )

    assert not found_messages, "Expected initial run messages to NOT appear in the logs."

    assert set(missing_messages) == set(
        RC_REMOTE_MESSAGES
    ), "The missing messages do not match the expected RC_REMOTE_MESSAGES."


@pytest.mark.skip(reason="Not implemented mock CDN")
@pytest.mark.parametrize(
    ("tcid", "error_message"),
    [
        pytest.param("LVPN-8452", "error: downloading main hash file:", id="no_cache"),
        pytest.param(
            "LVPN-8453", "failed downloading feature [nordvpn] remote config: downloading main file", id="no_config"
        ),
    ],
)
def test_remote_config_cdn_unavailable(
    tcid,  # noqa: ARG001
    error_message,  # noqa: ARG001
    initialized_app_with_remote_config,  # noqa: ARG001
    disable_remote_endpoint,  # noqa: ARG001
    set_custom_rc_retry_time_in_service,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :details    Verify that app handles the case, when remote files are not available in CDN and no cache files are locally

    :tcid       {tcid}

    :steps
        - # Start nordvpnd daemon
        - # Check daemon logs for error logs related to remote config
        - # Wait for next check
        - # Check daemon logs for error logs related to remote config

    :expected
        - # Nordvpn daemon shows error related to remote config
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows error related to remote config
    """
    cursor_before_restart = daemon_log_reader.get_cursor()
    daemon.restart()

    # Verification on start of application
    assert daemon_log_reader.wait_for_messages([error_message], cursor=cursor_before_restart, timeout=90), \
        "Couldn't found error logs"

    cursor_after_restart = daemon_log_reader.get_cursor()

    # Verification on retry mechanism
    assert daemon_log_reader.wait_for_messages([error_message], cursor=cursor_after_restart, timeout=90), \
        "Couldn't found error logs"


@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
def test_remote_config_download_config_on_start(
    initialized_app_with_remote_config,  # noqa: ARG001
    set_custom_rc_retry_time_in_service,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :details    Verify that app download config files and set file permissions to 600

    :tcid       LVPN-8456

    :steps
        - # Start nordvpnd daemon
        - # Check daemon logs for download logs related to remote config
        - # Wait for next check
        - # Check daemon logs for request logs related to remote config

    :expected
        - # Nordvpn daemon shows download related to remote config
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows request logs related to remote config
    """
    permission = "600"
    missed_feature_config = []
    daemon.restart()

    for message in RC_REMOTE_MESSAGES:
        if not daemon_log_reader.wait_for_messages([message], timeout=90):
            missed_feature_config.append(message)

    assert not missed_feature_config, f"Couldn't found download logs related to {missed_feature_config}"

    for message in RC_LOCAL_MESSAGES:
        if not daemon_log_reader.wait_for_messages([message], timeout=90):
            missed_feature_config.append(message)

    assert not missed_feature_config, f"Couldn't found download logs related to {missed_feature_config}"

    conf_files_data = sh.sudo.find(LOCAL_CACHE_DIR, "-type", "f", "-exec", "stat", "-c", "%a %y %n", "{}", ";")

    wrong_permissions_files = []
    for line in conf_files_data.stdout.decode("utf-8").splitlines():
        if permission not in line:
            wrong_permissions_files.append(line)

    assert not wrong_permissions_files, f"Found files that do not have {permission} chmod: {wrong_permissions_files}"

    check_log_for_request_get_messages(daemon_log_reader)


@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
@pytest.mark.parametrize(
    ("tcid", "is_config_different"),
    [
        pytest.param("LVPN-8477", False, id="equal_remote_config"),
        pytest.param(
            "LVPN-8500", True, id="different_remote_config", marks=pytest.mark.skip(reason="Not implemented mock CDN")
        ),
    ],
)
def test_remote_config_attempts_config(
    tcid,  # noqa: ARG001
    is_config_different,  # noqa: ARG001
    initialized_app_with_remote_config,  # noqa: ARG001
    set_custom_rc_retry_time_in_service,  # noqa: ARG001
    clean_cache_files,  # noqa: ARG001
    daemon_log_reader,
):
    """
    :details    Verify that app attempts to check for remote config files

    :tcid       {tcid}

    :steps
        - # Start nordvpnd daemon
        - # Check daemon log for download logs related to remote config
        - # Check date of config files
        - # Wait for next check
        - # Check daemon log for download logs related to remote config
        - # Check date time of config files

    :expected
        - # Nordvpn daemon shows download related to remote config
        - # Files are downloaded
        - # After 1 min, in next iteration, nordvpn daemon still checks and shows download logs related to remote config
        - # Files are downloaded/not downloaded according to remote config files
    """
    conf_files_data = sh.sudo.find.bake(LOCAL_CACHE_DIR, "-type", "f", "-exec", "stat", "-c", "%a", "%y", "%n", "{}", ";")
    chmod_to_be_found = "600"
    missed_local_config = []

    daemon.restart()

    for message in RC_LOCAL_MESSAGES:
        if not daemon_log_reader.wait_for_messages([message], timeout=90):
            missed_local_config.append(message)

    assert not missed_local_config, f"Couldn't found download logs related to {missed_local_config}"

    check_log_for_request_get_messages(daemon_log_reader)

    conf_files_data_before_attempt = conf_files_data().stdout.decode("utf-8").splitlines()

    wrong_permissions_files = []
    for line in conf_files_data_before_attempt:
        if chmod_to_be_found not in line:
            wrong_permissions_files.append(line)

    assert (
        not wrong_permissions_files
    ), f"Found files that do not have {chmod_to_be_found} chmod: {wrong_permissions_files}"

    check_log_for_request_get_messages(daemon_log_reader)

    conf_files_data_after_attempt = conf_files_data().stdout.decode("utf-8").splitlines()
    different_files = []

    if is_config_different:
        for nordvpn_file in ("nordvpn.json", "nordvpn-hash.json"):
            if nordvpn_file not in (
                list(set(conf_files_data_after_attempt) - set(conf_files_data_before_attempt))
            ):
                different_files.append(nordvpn_file)

    else:
        for line in conf_files_data_after_attempt:
            if line not in conf_files_data_before_attempt:
                different_files.append(line)

    assert not different_files, f"Found some files mismatch: {different_files}"


@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
@pytest.mark.parametrize(
    ("tcid", "parameter", "value", "verify_error_message", "error_message"),
    [
        pytest.param("LVPN-8544", "value", False, False, "", id="set_value_false"),
        pytest.param("LVPN-8522", "rollout", None, False, "", id="remove_rollout"),
        pytest.param("LVPN-8782", "app_version", "*", False, "", id="set_app_version_any"),
        pytest.param("LVPN-8518", "app_version", "", False, "", id="set_app_version_empty"),
        pytest.param(
            "LVPN-8519", "app_version", "y", True, RC_INVALID_VERSION_MESSAGE, id="set_app_version_invalid_version"
        ),
    ],
)
def test_remote_config_change_local_meshnet_config_settings(
    tcid,  # noqa: ARG001
    parameter,  # noqa: ARG001
    value,  # noqa: ARG001
    verify_error_message,
    error_message,
    backup_restore_rc_config_files,  # noqa: ARG001
    rc_config_manager,
    daemon_log_reader,  # noqa: ARG001
    set_custom_rc_retry_time_in_service,  # noqa: ARG001
    enable_local_config_in_service,  # noqa: ARG001
):
    """
    :details    Verify that application correctly responds to various remote configuration scenarios.

                Test that the app properly loads and applies remote configurations with different specifications
                for app versions and rollout settings, ensuring features activate appropriately in each case.

    :tcid       {tcid}

    :steps
        - # Change in meshnet.json file "value": false
        - # Change hash sum in meshnet-hash.json file
        - # Start nordvpnd daemon
        - # Check daemon for loading config logs
        - # Check that command "nordvpn meshnet" returns "Command 'meshnet' doesn't exist."

    :expected
        - # Value is set to False
        - # Config file was loaded by app
        - # Meshnet option is unavailable
    """
    if not rc_config_manager.set_permissions_cache_dir():
        pytest.skip("Directory doesn't exist")

    # Read current one meshnet config
    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}") as file:
        config = json.load(file)

    # Prepare config with changed value
    if value is not None:
        config["configs"][0]["settings"][0][parameter] = value
    elif value is None:
        del config["configs"][0]["settings"][0][parameter]

    # Open file with write permission and write new meshnet config
    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}", "w") as file:
        json.dump(config, file, indent=4)

    # Calculate new hash sum
    sha_sum_hash = sh.sha256sum(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_CONFIG_FILE}").stdout.decode("utf-8").split()[0]

    # Write new hash sum in meshnet hash file
    with open(f"{LOCAL_CACHE_DIR}/{RC_MESHNET_HASH_FILE}", "w") as file:
        json.dump({"hash": sha_sum_hash}, file)

    # Save a timestamp and restart daemon for picking up config
    cursor = daemon_log_reader.get_cursor()
    daemon.restart()

    assert daemon_log_reader.wait_for_messages(
        messages=[next(message for message in RC_LOCAL_MESSAGES if "meshnet" in message)], cursor=cursor, timeout=90
    ), "Couldn't found log of loading modified json file"

    # Additional verification for correct error message
    if verify_error_message:
        assert daemon_log_reader.wait_for_messages(
            messages=[error_message], cursor=cursor, timeout=90
        ), f"Couldn't found log {error_message}"

    # Verification if service works with incorrect config
    if value is False:
        with pytest.raises(sh.ErrorReturnCode):
            sh.nordvpn("meshnet")
    elif value is None:
        try:
            sh.nordvpn("meshnet")
        except sh.ErrorReturnCode:
            pytest.fail(reason="nordvpn meshnet is not enabled")


def test_killswitch_enabled_does_not_affect_cdn_with_firewall_mark(
        initialized_app_with_remote_config # noqa: ARG001
        ):
    """
    Test for a scenario where despite killswitch being enabled the network is still accessible for remote config.

    Test steps:
        1. Set up required environment variables
        2. Enable killswitch
        3. Remove previously fetched config files
        4. Attempt to fetch remote config
        5. Verify that the config is fetched successfully

    Jira ID: LVPN-8626
    """

    conf_dir = "/var/lib/nordvpn/conf/"
    expected_files = [
        "libtelio.json",
        "meshnet.json",
    ]
    def printdir():
        os.system(f"cd {conf_dir}/../ && ls -la && ls {conf_dir} -a")

    # enabling killswitch should not affect http transport of the remote config
    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    sh.nordvpn.disconnect()
    assert daemon.is_disconnected()
    print("--- before manual files removal")
    printdir()
    # remove previously fetched files
    # upon restart, they should be loaded again
    os.system(f"sudo rm -rf {conf_dir}")

    # make sure the rc files are gone
    for fname in expected_files:
        path = os.path.join(conf_dir, fname)
        res = os.popen(f"sudo test -f {path} && echo exists || echo missing").read().strip()
        assert res == "missing", f"File {path} should not exist"

    daemon.restart()
    print("--- after restart")
    printdir()

    for _ in range(10):
        time.sleep(1)
        printdir()

    for fname in expected_files:
        path = os.path.join(conf_dir, fname)
        res = os.popen(f"sudo test -f {path} && echo exists || echo missing").read().strip()
        assert res == "exists", f"File {os.path} should exist after kill-switch was enabled"


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_ech_disabled_for_not_allowed_technology(
        clean_cache_files,  # noqa: ARG001
        daemon_log_cursor: int,
        nordvpnd_scope_function,  # noqa: ARG001
        daemon_log_reader: LogReader,
        tech,
        proto,
        obfuscated,
):
    """
    Test to verify ECH is disabled for technologies that are not allowed.

    Test steps:
      1. Apply connection configuration for the parametrized combination:
         - technology
         - protocol
         - obfuscation
      2. Connect to VPN.
      3. Verify that the ECH message ('RC_USE_ECH_MESSAGE') is NOT present in the daemon log.
      4. Disconnect from VPN.

    :raises AssertionError: If the expected log message 'RC_USE_ECH_MESSAGE' is found in the daemon log
   """
    set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.connect()

    assert not daemon_log_reader.wait_for_messages(
        messages=RC_USE_ECH_MESSAGE, cursor=daemon_log_cursor
    ), f"Expected {RC_USE_ECH_MESSAGE} to be absent in the daemon log."

    sh.nordvpn.disconnect()


def test_ech_enabled_for_allowed_technology(
        clean_cache_files,  # noqa: ARG001
        daemon_log_cursor: int,
        nordvpnd_scope_function,  # noqa: ARG001
        daemon_log_reader: LogReader,
):
    """
    Test to verify ECH is enabled for technologies that are allowed(NordWhisper).

    Test steps:
      1. Apply connection configuration technology to NordWhisper.
      2. Connect to VPN.
      3. Verify that the ECH message ('RC_USE_ECH_MESSAGE') is present in the daemon log.
      4. Disconnect from VPN.

    :raises AssertionError: If the expected log message 'RC_USE_ECH_MESSAGE' is not found in the daemon log
   """
    set_technology_and_protocol("nordwhisper", "", "")

    sh.nordvpn.connect()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_USE_ECH_MESSAGE, cursor=daemon_log_cursor
    ), f"Expected {RC_USE_ECH_MESSAGE} to appear in the daemon log."

    sh.nordvpn.disconnect()


@pytest.mark.parametrize(
    "enable_ech",
    [
        True,
        False,
    ],
    ids=[
        "enabled",
        "disabled",
    ],
)
@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
def test_local_config_usage_ech_state(
        initialized_app_with_remote_config,  # noqa: ARG001
        rc_config_manager: RemoteConfigManager,
        daemon_log_reader: LogReader,
        enable_ech: bool,
):
    """
    Test to verify that when RC local usage is enabled, the app reads the local RC config and applies the ECH setting accordingly.

    Test steps:
      1. Set the local cached RC config to set 'enable_ech' to the parametrized value
         ('True' / 'False') and update the corresponding hash.
      2. Enable 'use local RC config' mode.
      3. Assert the daemon logs indicate local config usage ('RC_USE_LOCAL_CONFIG_MESSAGE').
      4. Set technology to NordWhisper.
      5. Connect to VPN.
      6. Assert ECH usage is logged ('RC_USE_ECH_MESSAGE') when enabled, and is absent when disabled.
      7. Disconnect from VPN and disable 'use local RC config' mode.

     :raises AssertionError: If the expected log message 'RC_USE_LOCAL_CONFIG_MESSAGE' is not found in the daemon log
     :raises AssertionError: If the expected log message 'RC_USE_ECH_MESSAGE' is not found in the daemon log if ECH
     is enabled and vice versa.
    """
    rc_config_manager.set_permissions_cache_dir()

    set_ech_for_local_config(enable_ech)

    with Defer(disable_rc_local_config_usage):
        enable_rc_local_config_usage()

        cursor = daemon_log_reader.get_cursor()

        daemon.restart()

        assert daemon_log_reader.wait_for_messages(
            messages=RC_USE_LOCAL_CONFIG_MESSAGE,
            cursor=cursor,
        )

        set_technology_and_protocol("nordwhisper", "", "")

        sh.nordvpn.connect()

        if enable_ech:
            assert daemon_log_reader.wait_for_messages(
                messages=RC_USE_ECH_MESSAGE, cursor=cursor
            ), f"Expected {RC_USE_ECH_MESSAGE} to be present in the daemon log."
        else:
            assert not daemon_log_reader.wait_for_messages(
                messages=RC_USE_ECH_MESSAGE, cursor=cursor
            ), f"Expected {RC_USE_ECH_MESSAGE} to be absent in the daemon log."

        sh.nordvpn.disconnect()


@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
def test_ech_disabled_with_local_rc_usage_then_reenabled_after_revert(
        initialized_app_with_remote_config,  # noqa: ARG001
        rc_config_manager: RemoteConfigManager,
        daemon_log_reader: LogReader,
):
    """
    Test to verify that ECH is disabled via local RC usage and restored after disabling local RC usage.

    Verify that disabling ECH via local RC config takes effect when local config usage is
    enabled, and that it reverts to normal RC config behavior once local config usage
    is disabled.

    Test steps:
      1. Set the local cached RC config to set 'enable_ech' to the 'False' and update the corresponding hash.
      2. Enable 'use local RC config' mode.
      3. Assert the daemon logs indicate local config usage ('RC_USE_LOCAL_CONFIG_MESSAGE').
      4. Set technology to NordWhisper.
      5. Connect to VPN.
      6. Assert ECH usage is NOT logged ('RC_USE_ECH_MESSAGE' absent).
      7. Disconnect from VPN.
      8. Disable 'use local RC config' mode.
      9. Restart the daemon.
     10. Assert the daemon logs no longer indicate local config usage ('RC_USE_LOCAL_CONFIG_MESSAGE' absent).
     11. Connect to VPN.
     12. Assert ECH usage IS logged ('RC_USE_ECH_MESSAGE' present).
     13. Disconnect from VPN.

     :raises AssertionError: If the expected log message 'RC_USE_LOCAL_CONFIG_MESSAGE' is not found in the daemon log
     :raises AssertionError: If the expected log message 'RC_USE_ECH_MESSAGE' is found in the daemon log if ECH
     is disabled and vice versa when RC local config usage is disabled.
    """
    rc_config_manager.set_permissions_cache_dir()

    set_ech_for_local_config(False)

    with Defer(disable_rc_local_config_usage):
        enable_rc_local_config_usage()

        cursor = daemon_log_reader.get_cursor()

        daemon.restart()

        assert daemon_log_reader.wait_for_messages(
            messages=RC_USE_LOCAL_CONFIG_MESSAGE,
            cursor=cursor,
        )

        set_technology_and_protocol("nordwhisper", "", "")

        sh.nordvpn.connect()

        assert not daemon_log_reader.wait_for_messages(
            messages=RC_USE_ECH_MESSAGE, cursor=cursor
        ), f"Expected {RC_USE_ECH_MESSAGE} to be absent in the daemon log."

        sh.nordvpn.disconnect()

    cursor = daemon_log_reader.get_cursor()

    daemon.restart()

    assert not daemon_log_reader.wait_for_messages(
        messages=RC_USE_LOCAL_CONFIG_MESSAGE,
        cursor=cursor,
    )

    sh.nordvpn.connect()

    assert daemon_log_reader.wait_for_messages(
        messages=RC_USE_ECH_MESSAGE, cursor=cursor
    ), f"Expected {RC_USE_ECH_MESSAGE} to be present in the daemon log after reverting."

    sh.nordvpn.disconnect()


@pytest.mark.parametrize(
    "enable_ech",
    [
        True,
        False,
    ],
    ids=[
        "value_enabled",
        "value_disabled",
    ],
)
@pytest.mark.skipif(daemon.get_env() == Env.PROD, reason="Not applicable in prod environment")
def test_local_config_usage_missing_enable_ech_field(
        initialized_app_with_remote_config,  # noqa: ARG001
        rc_config_manager: RemoteConfigManager,
        daemon_log_reader: LogReader,
        enable_ech: bool,
):
    """
    Test to verify app behavior when 'enable_ech' is removed from the local RC config while local RC usage is enabled.

    Verify that removing the 'enable_ech' entry from the local cached RC nordwhisper config takes effect when local
    config usage is enabled, and that the daemon logs a warning about the missing parameter during NordWhisper feature
    fetch without breaking the connection flow.

    Test steps:
      1. Update the local cached RC config by removing the '("name", "enable_ech")' field from the matching entry and
         update the corresponding hash.
      2. Enable 'use local RC config' mode.
      3. Restart the daemon.
      4. Assert the daemon logs indicate local config usage ('RC_USE_LOCAL_CONFIG_MESSAGE').
      5. Set technology to NordWhisper.
      6. Connect to VPN.
      7. Assert the daemon log contains the missing parameter warning:
         '[Warning] Failed to fetch NordWhisper features: feature [nordwhisper] param [enable_ech] not found'.
      8. Assert ECH usage IS logged ('RC_USE_ECH_MESSAGE' present).
      9. Disconnect from VPN.
      10. Disable 'use local RC config' mode.

    :raises AssertionError: If the expected log message 'RC_USE_LOCAL_CONFIG_MESSAGE' is not found in the daemon log.
    :raises AssertionError: If the expected missing-parameter warning message is not found in the daemon log.
    :raises AssertionError: If the expected log message 'RC_USE_ECH_MESSAGE' is not found in the daemon log
    """
    rc_config_manager.set_permissions_cache_dir()

    set_ech_for_local_config(enable_ech, entries_to_remove=[("name", "enable_ech")])

    with Defer(disable_rc_local_config_usage):
        enable_rc_local_config_usage()

        cursor = daemon_log_reader.get_cursor()

        daemon.restart()

        assert daemon_log_reader.wait_for_messages(
            messages=RC_USE_LOCAL_CONFIG_MESSAGE,
            cursor=cursor,
        )

        set_technology_and_protocol("nordwhisper", "", "")

        sh.nordvpn.connect()

        missing_ech_field_message = \
            "[Warning] Failed to fetch NordWhisper features: feature [nordwhisper] param [enable_ech] not found"

        assert daemon_log_reader.wait_for_messages(
            messages=[missing_ech_field_message], cursor=cursor
        ), f"Expected {missing_ech_field_message} to be present in the daemon log."

        assert daemon_log_reader.wait_for_messages(
            messages=RC_USE_ECH_MESSAGE, cursor=cursor
        ), f"Expected {RC_USE_ECH_MESSAGE} to be present in the daemon log."

        sh.nordvpn.disconnect()

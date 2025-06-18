import json
import os
import io
import pexpect

import sh

from . import UserConsentMode, logging, ssh, squash_whitespace, WE_VALUE_YOUR_PRIVACY_MSG, USER_CONSENT_POMPT


class Credentials:
    def __init__(self, email, token, password):
        self.email = email
        self.token = token
        self.password = password


def get_credentials(key) -> Credentials:
    """Returns token by a given key."""
    na_credentials = os.environ.get("NA_TESTS_CREDENTIALS")
    na_credentials_key = os.environ.get("NA_CREDENTIALS_KEY")
    full_key = key if na_credentials_key is None else f"{key}_{na_credentials_key}"

    if na_credentials is None:
        raise Exception("environment variable 'NA_TESTS_CREDENTIALS' is not set")
    creds = json.loads(na_credentials)

    key = key if creds.get(full_key) is None else full_key

    creds = creds[key]

    return Credentials(
            email=creds.get("email", None),
            token=creds.get("token", None),
            password=creds.get("password", None))


def login_as(username, ssh_client: ssh.Ssh = None, with_user_consent: UserConsentMode = UserConsentMode.ENABLED):
    """login_as specified user, optional SSH connection and option for setting user consent before calling login."""
    token = get_credentials(username).token

    logging.log(f"logging in as {token}")

    if ssh_client is not None:
        if with_user_consent != UserConsentMode.UNDEFINED:
            ssh_client.exec_command(f"nordvpn set analytics {_analytics_value(with_user_consent)}")
        return ssh_client.exec_command(f"nordvpn login --token {token}")

    if with_user_consent != UserConsentMode.UNDEFINED:
        sh.nordvpn.set.analytics(_analytics_value(with_user_consent))
    return sh.nordvpn.login("--token", token)


def _analytics_value(mode: UserConsentMode) -> str:
    if mode == UserConsentMode.UNDEFINED:
        raise Exception("can't set analytics with undefined consent")

    if mode == UserConsentMode.ENABLED:
        return "on"

    if mode == UserConsentMode.DISABLED:
        return "off"

    msg = f"not supported consent mode: {mode}"
    raise Exception(msg)


def spawn_nordvpn_login():
    """Spawns the nordvpn login process and sets up an output buffer."""
    buffer = io.StringIO()
    cli = pexpect.spawn("nordvpn", args=["login"], encoding="utf-8", timeout=10)
    cli.logfile_read = buffer
    return cli, buffer


def wait_for_consent_prompt(cli):
    """Waits for the consent prompt to appear."""
    cli.expect(USER_CONSENT_POMPT)


def get_new_output(buffer, old_output=""):
    """Returns only the newly added output since old_output."""
    return buffer.getvalue()[len(old_output):]


def assert_prompt_present(output, message=WE_VALUE_YOUR_PRIVACY_MSG):
    """Asserts that the consent message is in the output."""
    assert message in squash_whitespace(output)


def assert_prompt_absent(output, message=WE_VALUE_YOUR_PRIVACY_MSG):
    """Asserts that the consent message is not in the output."""
    assert message not in squash_whitespace(output)

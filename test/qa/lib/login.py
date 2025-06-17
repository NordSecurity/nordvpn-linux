import json
import os
from enum import Enum

import sh

from . import logging, ssh


class ConsentMode(Enum):
    Undefined = 0
    Granted = 1
    Denied = 2


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


def login_as(username, ssh_client: ssh.Ssh = None, with_analytics: ConsentMode = ConsentMode.Granted):
    """login_as specified user with optional delay before calling login."""
    token = get_credentials(username).token

    logging.log(f"logging in as {token}")

    if ssh_client is not None:
        if with_analytics != ConsentMode.Undefined:
            ssh_client.exec_command(f"nordvpn set analytics {analytics_value(with_analytics)}")
        return ssh_client.exec_command(f"nordvpn login --token {token}")

    if with_analytics != ConsentMode.Undefined:
        sh.nordvpn.set.analytics(analytics_value(with_analytics))
    return sh.nordvpn.login("--token", token)


def analytics_value(mode: ConsentMode) -> str:
    if mode == ConsentMode.Undefined:
        raise "can't enable analytics with undefined consent"
    elif mode == ConsentMode.Granted:
        return "on"
    elif mode == ConsentMode.Denied:
        return "off"
    raise f"not supported consent mode: {mode}"


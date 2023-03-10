from lib import logging, ssh
import sh
import time
import json
import os

def get_default_credentials():
    """returns tuple[username,token]"""
    default_username = os.environ.get("DEFAULT_LOGIN_USERNAME")
    default_token = os.environ.get("DEFAULT_LOGIN_TOKEN")
    ci_credentials = os.environ.get("NA_TESTS_CREDENTIALS")
    if ci_credentials is not None:
        devs = json.loads(ci_credentials)
        dev_email = os.environ.get("GITLAB_USER_EMAIL")
        if dev_email in devs:
            default_username = devs[dev_email]["username"]
            default_token= devs[dev_email]["token"]
    return default_username, default_token

def login_as(username, ssh_client: ssh.Ssh = None):
    """login_as specified user with optional delay before calling login"""

    default_username, default_token = get_default_credentials()
    users = {
        "default": [
            default_username,
            default_token,
        ],
        "invalid": [
            os.environ.get("DEFAULT_LOGIN_USERNAME"),
            "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        ],
        "valid": [
            os.environ.get("VALID_LOGIN_USERNAME"),
            os.environ.get("VALID_LOGIN_TOKEN"),
        ],
        "expired": [
            os.environ.get("EXPIRED_LOGIN_USERNAME"),
            os.environ.get("EXPIRED_LOGIN_TOKEN"),
        ],
        "qa-peer": [
            os.environ.get("QA_PEER_USERNAME"),
            os.environ.get("QA_PEER_TOKEN"),
        ],
    }

    user = users[username]
    logging.log(f"logging in as {user[0]}")

    if ssh_client is not None:
        return ssh_client.exec_command(f"nordvpn login --token {user[1]}")
    else:
        return sh.nordvpn.login("--token", user[1])

import pytest
import sh

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)


def setup_module(module):  # noqa: ARG001
    daemon.start()


def teardown_module(module):  # noqa: ARG001
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


def test_login():
    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        output = login.login_as("default")
        assert "Welcome to NordVPN! You can now connect to VPN by using 'nordvpn connect'." in output


def test_invalid_token_login():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--token", "xyz%#")

    assert "We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support." in str(ex.value)


def test_repeated_login():
    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            login.login_as("default")

        assert "You are already logged in." in str(ex.value)


@pytest.mark.skip(reason="can't get login token for expired account")
def test_expired_account_connect():
    lib.set_technology_and_protocol("openvpn", "udp", "off")

    output = login.login_as("expired")
    print(output)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect()

    assert "Your account has expired." in str(ex.value)
    assert "https://join.nordvpn.com/order/?utm_medium=app&utm_source=linux" in str(
        ex.value
    )
    sh.nordvpn.logout("--persist-token")


def test_login_while_connected():
    login.login_as("default")

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
            with pytest.raises(sh.ErrorReturnCode_1) as ex:
                login.login_as("valid")

        assert "You are already logged in." in str(ex.value)

    assert network.is_disconnected()


def test_login_without_internet():
    default_gateway = network.stop()

    with pytest.raises(sh.ErrorReturnCode_1):
        login.login_as("default")

    network.start(default_gateway)


def test_repeated_logout():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.logout("--persist-token")

    assert "You are not logged in." in str(ex.value)


def test_logged_out_connect():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect()

    assert "You are not logged in." in str(ex.value)


def test_logout_disconnects():
    output = login.login_as("default")
    print(output)

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)

    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.logout("--persist-token")
    print(output)
    assert "You are logged out." in output
    assert network.is_disconnected()


def test_missing_token_login():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--token")

    assert "Token parameter value is missing." in str(ex.value)


def test_missing_url_callback_login():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--callback")

    assert "Expected a url." in str(ex.value)


def test_invalid_url_callback_login():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--callback", "https://www.google.com/")

    assert "Expected a url with nordvpn scheme." in str(ex.value)


def test_repeated_login_callback():
    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback")

        assert "You are already logged in." in str(ex.value)


def test_repeated_login_callback_invalid_url():
    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback", "https://www.google.com/")

        assert "You are already logged in." in str(ex.value)


def test_repeated_login_callback_nordvpn_scheme_url():
    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback", "nordvpn://")

        assert "You are already logged in." in str(ex.value)

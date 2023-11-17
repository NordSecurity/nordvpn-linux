from lib import (
    daemon,
    info,
    logging,
    login,
    network,
)
import lib
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()


def teardown_module(module):
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


def test_login():
    output = login.login_as("default")
    assert "Welcome to NordVPN! You can now connect to VPN by using 'nordvpn connect'." in output
    sh.nordvpn.logout("--persist-token")


def test_invalid_token_login():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--token", "xyz%#")
    
    assert "We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support." in str(ex.value)


def test_repeated_login():
    output = login.login_as("default")
    print(output)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        login.login_as("default")

    assert "You are already logged in." in str(ex.value)
    sh.nordvpn.logout("--persist-token")

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


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_login_while_connected():
    output = login.login_as("default")
    print(output)

    output = sh.nordvpn.connect()
    print(output)

    assert lib.is_connect_successful(output)
    assert network.is_connected()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        login.login_as("valid")

    assert "You are already logged in." in str(ex.value)

    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
    sh.nordvpn.logout("--persist-token")


def test_login_without_internet():
    default_gateway = network.stop()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
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


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_logout_disconnects():
    output = login.login_as("default")
    print(output)

    output = sh.nordvpn.connect()
    print(output)

    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.logout("--persist-token")
    print(output)
    assert "You are logged out." in output
    assert network.is_disconnected()


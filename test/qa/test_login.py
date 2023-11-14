import os

import pytest
import sh
import timeout_decorator

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


@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
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


@pytest.mark.skip("Does not work on Docker")
@pytest.mark.parametrize("login_flag", login.LOGIN_FLAG)
#@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_selenium_login(login_flag):
    preferences = [
        ["network.protocol-handler.expose.nordvpn", True],
        ["network.protocol-handler.external.nordvpn", True],
        ["network.protocol-handler.app.nordvpn", "/usr/bin/nordvpn"]
        ]

    selenium = login.SeleniumBrowser(preferences)
    browser = selenium.browser_get()

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with lib.Defer(selenium.browser_kill):
            # Get login link from NordVPN app, trim all spaces & chars after link itself
            login_link = sh.nordvpn.login(login_flag).split(": ")[1][:login.LOGIN_LINK_LENGTH]

            # Open login link from NordVPN app
            browser.get(login_link)

            # User credentials, that we will use in order to log in to NordAccount
            user_info = os.environ.get("DEFAULT_LOGIN_USERNAME") + ":" + os.environ.get("DEFAULT_LOGIN_PASSWORD")

            try:
                # Username page
                selenium.browser_element_interact(login.NA_USERNAME_PAGE_TEXTBOX_XPATH, user_info.split(':')[0])
                selenium.browser_element_interact(login.NA_USERNAME_PAGE_BUTTON_XPATH)

                # Password page
                selenium.browser_element_interact(login.NA_PASSWORD_PAGE_TEXTBOX_XPATH, user_info.split(':')[1])
                selenium.browser_element_interact(login.NA_PASSWORD_PAGE_BUTTON_XPATH)

                # Continue to app page
                selenium.browser_element_interact(login.NA_CONTINUE_PAGE_LINK_BUTTON)
            except:  # noqa: E722
                browser.save_screenshot(login.BROWSER_LOGS_PATH + "Screenshot.png")
                pytest.fail()

            assert login.LOGOUT_MSG_SUCCESS in sh.nordvpn.logout()


@pytest.mark.parametrize("login_flag", login.LOGIN_FLAG)
#@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_selenium_login_callback(login_flag):
    selenium = login.SeleniumBrowser()
    browser = selenium.browser_get()

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with lib.Defer(selenium.browser_kill):
            # Get login link from NordVPN app, trim all spaces & chars after link itself
            login_link = sh.nordvpn.login(login_flag).split(": ")[1][:login.LOGIN_LINK_LENGTH]

            # Open login link from NordVPN app
            browser.get(login_link)

            # User credentials, that we will use in order to log in to NordAccount
            user_info = os.environ.get("DEFAULT_LOGIN_USERNAME") + ":" + os.environ.get("DEFAULT_LOGIN_PASSWORD")

            try:
                # Username page
                selenium.browser_element_interact(login.NA_USERNAME_PAGE_TEXTBOX_XPATH, user_info.split(':')[0])
                selenium.browser_element_interact(login.NA_USERNAME_PAGE_BUTTON_XPATH)

                # Password page
                selenium.browser_element_interact(login.NA_PASSWORD_PAGE_TEXTBOX_XPATH, user_info.split(':')[1])
                selenium.browser_element_interact(login.NA_PASSWORD_PAGE_BUTTON_XPATH)

                # Continue to app page
                # preferences not set in constructor, so when we click link it does not redirect us to app.
                callback_link = selenium.browser_element_interact(login.NA_CONTINUE_PAGE_LINK_BUTTON, return_attribute='href')
            except:  # noqa: E722
                browser.save_screenshot(login.BROWSER_LOGS_PATH + "Screenshot.png")
                pytest.fail()

            assert login.LOGIN_MSG_SUCCESS in sh.nordvpn.login("--callback", callback_link)

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
)


def setup_function(function):  # noqa: ARG001
    daemon.start()

    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.set.defaults()
    daemon.stop()


@pytest.mark.skip("Does not work on Docker")
@pytest.mark.parametrize("login_flag", login.LOGIN_FLAG)
@timeout_decorator.timeout(60)
def test_selenium_login(login_flag):
    preferences = [
        ["network.protocol-handler.expose.nordvpn", True],
        ["network.protocol-handler.external.nordvpn", True],
        ["network.protocol-handler.app.nordvpn", "/usr/bin/nordvpn"]
        ]

    selenium = login.SeleniumBrowser(preferences)
    browser = selenium.browser_get()

    with lib.Defer(selenium.browser_kill):
        # Get login link from NordVPN app, trim all spaces & chars after link itself
        login_link = sh.nordvpn.login(login_flag, _tty_out=False).strip().split(": ")[1]

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
@timeout_decorator.timeout(60)
def test_selenium_login_callback(login_flag):
    selenium = login.SeleniumBrowser()
    browser = selenium.browser_get()

    with lib.Defer(selenium.browser_kill):
        # Get login link from NordVPN app, trim all spaces & chars after link itself
        login_link = sh.nordvpn.login(login_flag, _tty_out=False).strip().split(": ")[1]

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

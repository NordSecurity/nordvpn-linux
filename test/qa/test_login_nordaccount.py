import pytest
import sh

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    selenium,
)


def setup_function(function):  # noqa: ARG001
    daemon.start()

    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.set.defaults("--logout")
    daemon.stop()


@pytest.mark.skip("Does not work on Docker")
@pytest.mark.parametrize("login_flag", selenium.LOGIN_FLAG)
def test_selenium_login(login_flag):
    preferences = [
        ["network.protocol-handler.expose.nordvpn", True],
        ["network.protocol-handler.external.nordvpn", True],
        ["network.protocol-handler.app.nordvpn", "/usr/bin/nordvpn"]
        ]

    sel = selenium.SeleniumBrowser(preferences)
    browser = sel.browser_get()

    with lib.Defer(sel.browser_kill):
        # Get login link from NordVPN app, trim all spaces & chars after link itself
        login_link = sh.nordvpn.login(login_flag, _tty_out=False).strip().split(": ")[1]
        print(f"Login link: {login_link}\n")

        # Open login link from NordVPN app
        browser.get(login_link)
        print(f"Browser URL: {browser.current_url}\n")

        try:
            credentials = login.get_credentials("default")
            # Username page
            sel.browser_element_interact(selenium.NA_USERNAME_PAGE_TEXTBOX_XPATH, credentials.email)
            sel.browser_element_interact(selenium.NA_USERNAME_PAGE_BUTTON_XPATH)

            # Password page
            sel.browser_element_interact(selenium.NA_PASSWORD_PAGE_TEXTBOX_XPATH, credentials.password)
            sel.browser_element_interact(selenium.NA_PASSWORD_PAGE_BUTTON_XPATH)

            # Continue to app page
            sel.browser_element_interact(selenium.NA_CONTINUE_PAGE_LINK_BUTTON)
        except:  # noqa: E722
            browser.save_screenshot(selenium.BROWSER_LOGS_PATH + "Screenshot.png")
            pytest.fail()

        output = sh.nordvpn.logout(_tty_out=False)
        print(f"Logout action output: {output}\n")
        assert selenium.LOGOUT_MSG_SUCCESS in output


@pytest.mark.parametrize("login_flag", selenium.LOGIN_FLAG)
def test_selenium_login_callback(login_flag):
    sel = selenium.SeleniumBrowser()
    browser = sel.browser_get()

    with lib.Defer(sel.browser_kill):
        sh.nordvpn.set.analytics("on")
        # Get login link from NordVPN app, trim all spaces & chars after link itself
        login_link = sh.nordvpn.login(login_flag, _tty_out=False).strip().split(": ")[1]
        print(f"Login link: {login_link}\n")

        # Open login link from NordVPN app
        browser.get(login_link)
        print(f"Browser URL: {browser.current_url}\n")

        try:
            credentials = login.get_credentials("default")
            # Username page
            sel.browser_element_interact(selenium.NA_USERNAME_PAGE_TEXTBOX_XPATH, credentials.email)
            sel.browser_element_interact(selenium.NA_USERNAME_PAGE_BUTTON_XPATH)

            # Password page
            sel.browser_element_interact(selenium.NA_PASSWORD_PAGE_TEXTBOX_XPATH, credentials.password)
            sel.browser_element_interact(selenium.NA_PASSWORD_PAGE_BUTTON_XPATH)

            # Continue to app page
            # preferences not set in constructor, so when we click link it does not redirect us to app.
            callback_link = sel.browser_element_interact(selenium.NA_CONTINUE_PAGE_LINK_BUTTON, return_attribute='href')
            print(f"Callback URL: {callback_link}\n")
        except:  # noqa: E722
            browser.save_screenshot(selenium.BROWSER_LOGS_PATH + "Screenshot.png")
            pytest.fail()

        output = sh.nordvpn.login("--callback", callback_link, _tty_out=False)
        print(f"Callback login action output: {output}\n")
        assert selenium.LOGIN_MSG_SUCCESS in output

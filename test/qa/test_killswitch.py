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
    login.login_as("default")


def teardown_module(module):  # noqa: ARG001
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):  # noqa: ARG001
    logging.log()


def teardown_function(function):  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()


MSG_KILLSWITCH_ON = "Kill Switch is set to 'enabled' successfully."
MSG_KILLSWITCH_OFF = "Kill Switch is set to 'disabled' successfully."


@pytest.mark.parametrize("login_flag", login.LOGIN_FLAG)
#@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(60)
def test_a(login_flag):
    sh.nordvpn.logout("--persist-token")

    selenium = login.SeleniumBrowser()
    browser = selenium.browser_get()

    #with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
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
            pytest.fail("Exception occured")

        assert login.LOGIN_MSG_SUCCESS in sh.nordvpn.login("--callback", callback_link)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_killswitch_on_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_on_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_on_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    with lib.Defer(sh.nordvpn.disconnect):
        output = sh.nordvpn.connect()
        print(output)
        assert network.is_connected()

        assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
        assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_off_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.Defer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

            assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
            assert not daemon.is_killswitch_on()
            assert network.is_available()

    assert network.is_available()


@pytest.mark.parametrize(("tech_from", "proto_from", "obfuscated_from"), lib.TECHNOLOGIES)
@pytest.mark.parametrize(("tech_to", "proto_to", "obfuscated_to"), lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_killswitch_reconnect(tech_from, proto_from, obfuscated_from, tech_to, proto_to, obfuscated_to):
    lib.set_technology_and_protocol(tech_from, proto_from, obfuscated_from)
    assert network.is_available()

    assert MSG_KILLSWITCH_ON in sh.nordvpn.set.killswitch("on")
    assert daemon.is_killswitch_on()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

        with lib.ErrorDefer(sh.nordvpn.disconnect):
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

            lib.set_technology_and_protocol(tech_to, proto_to, obfuscated_to)
            assert network.is_connected()
            output = sh.nordvpn.connect()
            print(output)
            assert network.is_connected()

    output = sh.nordvpn.disconnect()
    print(output)
    assert daemon.is_disconnected()

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        assert network.is_not_available(2)

    assert MSG_KILLSWITCH_OFF in sh.nordvpn.set.killswitch("off")
    assert not daemon.is_killswitch_on()
    assert network.is_available()


# Test for 3.8.7 hotfix. Account and login commands would not work when killswitch is on
# Issue 441
def test_fancy_transport():
    sh.nordvpn.logout("--persist-token")
    output = sh.nordvpn.set.killswitch("on")
    assert MSG_KILLSWITCH_ON in output

    output = login.login_as("default")
    print(output)
    assert "Welcome to NordVPN!" in output

    with lib.ErrorDefer(sh.nordvpn.set.killswitch.off):
        output = sh.nordvpn.account()
        print(output)
        assert "Account Information:" in output

    sh.nordvpn.set.killswitch("off")
    assert network.is_available()

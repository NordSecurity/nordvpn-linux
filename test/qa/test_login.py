import pytest
import pexpect
import io
import sh

import lib
from lib import (
    daemon,
    info,
    logging,
    login,
    network,
    settings,
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


def test_analytics_consent_is_displayed_on_login():
    cli = pexpect.spawn("nordvpn", args=["login"], encoding='utf-8', timeout=10)
    # wait until the final prompt appears, then capture everything before it
    cli.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")
    full_output = cli.before + cli.after  # everything printed so far, including the last line

    assert (
        lib.squash_whitespace(lib.EXPECTED_CONSENT_MESSAGE)
        in lib.squash_whitespace(full_output)
    ), "Consent message did not match expected full output"


def test_invalid_input_repeats_consent_prompt_only():
    output_buffer = io.StringIO()
    cli = pexpect.spawn("nordvpn", args=["login"], encoding="utf-8", timeout=10)
    cli.logfile_read = output_buffer

    # wait for first full prompt
    cli.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")
    first_output = output_buffer.getvalue()

    # send invalid input
    cli.sendline("blah")

    # wait for error + prompt again
    cli.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")
    second_output = output_buffer.getvalue()[len(first_output):]  # take just the new part

    assert "We value your privacy" in lib.squash_whitespace(first_output)
    assert "We value your privacy" not in lib.squash_whitespace(second_output)
    assert "Invalid response" in lib.squash_whitespace(second_output)
    assert "(y/n)" in lib.squash_whitespace(second_output)


def test_analytics_consent_prompt_reappears_after_ctrl_c_interrupt():
    # first run: user is interrupted with Ctrl+C at the prompt
    cli1 = pexpect.spawn("nordvpn", args=["login"], encoding="utf-8", timeout=10)
    buffer1 = io.StringIO()
    cli1.logfile_read = buffer1

    # wait for the consent prompt
    cli1.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")

    # simulate user hitting Ctrl+C
    cli1.sendintr()  # this sends SIGINT like Ctrl+C
    cli1.expect(pexpect.EOF)
    first_output = buffer1.getvalue()

    assert "We value your privacy" in lib.squash_whitespace(first_output), \
        "Consent prompt not shown before Ctrl+C"

    # second run: should still see the prompt
    cli2 = pexpect.spawn("nordvpn", args=["login"], encoding="utf-8", timeout=10)
    buffer2 = io.StringIO()
    cli2.logfile_read = buffer2

    cli2.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")
    second_output = buffer2.getvalue()

    assert "We value your privacy" in lib.squash_whitespace(second_output), \
        "Consent prompt did not reappear on second login attempt"


def test_analytics_consent_granted_after_pressing_y_and_does_not_appear_again():
    # first run: prompt appears, user consents
    cli = pexpect.spawn("nordvpn", args=["login"], encoding='utf-8', timeout=10)
    cli.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)")

    assert not settings.is_analytics_consent_declared(), "Consent should not be declared before interaction"

    cli.sendline("y")
    cli.expect(pexpect.EOF)

    assert settings.is_analytics_consent_granted(), "Consent should be recorded after pressing 'y'"

    # second run: ensure the consent prompt does NOT appear again
    cli2 = pexpect.spawn("nordvpn", args=["login"], encoding='utf-8', timeout=10)

    try:
        # try to match the consent prompt
        cli2.expect(r"Do you allow us to collect and use limited app performance data\? \(y/n\)", timeout=3)
        assert False, "Consent prompt appeared again after it was already granted"
    except pexpect.exceptions.TIMEOUT:
        # good - the prompt didn't appear
        pass
    except pexpect.exceptions.EOF:
        # also good — the process exited before showing the prompt
        pass

    cli2.expect(pexpect.EOF)


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

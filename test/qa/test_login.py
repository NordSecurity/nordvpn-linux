import pytest
import pexpect
import sh

import lib
from lib import (
    daemon,
    login,
    network,
    settings,
)


pytestmark = pytest.mark.usefixtures("collect_logs")


def setup_module(module):  # noqa: ARG001
    daemon.start()


def teardown_module(module):  # noqa: ARG001
    daemon.stop()


def test_user_consent_is_displayed_on_login():
    """Manual TC: LVPN-8241"""

    cli, _ = login.spawn_nordvpn_login()
    login.wait_for_consent_prompt(cli)

    full_output = cli.before + cli.after  # everything printed so far, including the last line

    assert (
        lib.squash_whitespace(lib.EXPECTED_CONSENT_MESSAGE)
        in lib.squash_whitespace(full_output)
    ), "Consent message did not match expected full output"


def test_invalid_input_repeats_consent_prompt_only():
    """Manual TC: LVPN-8227"""

    cli, buffer = login.spawn_nordvpn_login()
    login.wait_for_consent_prompt(cli)
    first_output = buffer.getvalue()

    cli.sendline("blah")
    login.wait_for_consent_prompt(cli)
    second_output = login.get_new_output(buffer, first_output)

    login.assert_prompt_present(first_output)
    login.assert_prompt_absent(second_output)
    assert "Invalid response" in lib.squash_whitespace(second_output)
    assert "(y/n)" in lib.squash_whitespace(second_output)


def test_user_consent_prompt_reappears_after_ctrl_c_interrupt():
    """Manual TC: LVPN-8244"""

    cli1, buffer1 = login.spawn_nordvpn_login()
    login.wait_for_consent_prompt(cli1)
    cli1.sendintr()
    cli1.expect(pexpect.EOF)
    output1 = buffer1.getvalue()
    login.assert_prompt_present(output1)

    cli2, buffer2 = login.spawn_nordvpn_login()
    login.wait_for_consent_prompt(cli2)
    output2 = buffer2.getvalue()
    login.assert_prompt_present(output2)


def test_user_consent_granted_after_pressing_y_and_does_not_appear_again():
    """Manual TC: LVPN-8772"""

    cli, _ = login.spawn_nordvpn_login()
    login.wait_for_consent_prompt(cli)

    assert not settings.is_user_consent_declared(), "Consent should not be declared before interaction"
    cli.sendline("y")
    cli.expect(pexpect.EOF)
    assert settings.is_user_consent_granted(), "Consent should be recorded after pressing 'y'"

    cli2, _ = login.spawn_nordvpn_login()
    try:
        cli2.expect(lib.USER_CONSENT_PROMPT)
        raise AssertionError("Consent prompt appeared again after it was already granted")
    except (pexpect.exceptions.TIMEOUT, pexpect.exceptions.EOF):
        pass  # Good, prompt did not appear
    cli2.expect(pexpect.EOF)


def test_login():
    """Manual TC: LVPN-505"""

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        output = login.login_as("default")
        assert "Welcome to NordVPN! You can now connect to the VPN by using 'nordvpn connect'." in output


def test_invalid_token_login():
    """Manual TC: LVPN-503"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--token", "xyz%#")

    assert "We couldn't log you in - the access token is not valid. Please check if you've entered the token correctly. If the issue persists, contact our customer support." in ex.value.stdout.decode("utf-8")


def test_repeated_login():
    """Manual TC: LVPN-549"""

    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            login.login_as("default")

        assert "You're already logged in." in ex.value.stdout.decode("utf-8")


@pytest.mark.skip(reason="can't get login token for expired account")
def test_expired_account_connect():
    """Manual TC: LVPN-1407"""

    lib.set_technology_and_protocol("openvpn", "udp", "off")

    output = login.login_as("expired")
    print(output)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect()

    assert "Your account has expired." in ex.value.stdout.decode("utf-8")
    assert "https://join.nordvpn.com/order/?utm_medium=app&utm_source=linux" in ex.value.stdout.decode("utf-8")
    sh.nordvpn.logout("--persist-token")


def test_login_while_connected():
    """Manual TC: LVPN-8577"""

    login.login_as("default")

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
            with pytest.raises(sh.ErrorReturnCode_1) as ex:
                login.login_as("valid")

        assert "You're already logged in" in ex.value.stdout.decode("utf-8")

    assert network.is_disconnected()


def test_login_without_internet():
    """Manual TC: LVPN-8578"""

    default_gateway = network.stop()

    with pytest.raises(sh.ErrorReturnCode_1):
        login.login_as("default")

    network.start(default_gateway)


def test_repeated_logout():
    """Manual TC: LVPN-695"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.logout("--persist-token")

    assert "You're not logged in" in ex.value.stdout.decode("utf-8")


def test_logged_out_connect():
    """Manual TC: LVPN-8690"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect()

    assert "You're not logged in" in ex.value.stdout.decode("utf-8")


def test_logout_disconnects():
    """Manual TC: LVPN-8752"""

    output = login.login_as("default")
    print(output)

    output = sh.nordvpn.connect(_tty_out=False)
    print(output)

    assert lib.is_connect_successful(output)
    assert network.is_connected()

    output = sh.nordvpn.logout("--persist-token")
    print(output)
    assert "You're logged out." in output
    assert network.is_disconnected()


def test_missing_token_login():
    """Manual TC: LVPN-485"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--token")

    assert "Token parameter value is missing." in ex.value.stdout.decode("utf-8")


def test_missing_url_callback_login():
    """Manual TC: LVPN-667"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--callback")

    assert "Expected a url." in ex.value.stdout.decode("utf-8")


def test_invalid_url_callback_login():
    """Manual TC: LVPN-472"""

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.login("--callback", "https://www.google.com/")

    assert "Expected a url with nordvpn scheme." in ex.value.stdout.decode("utf-8")


def test_repeated_login_callback():
    """Manual TC: LVPN-443"""

    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback")

        assert "You're already logged in." in ex.value.stdout.decode("utf-8")


def test_repeated_login_callback_invalid_url():
    """Manual TC: LVPN-439"""

    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback", "https://www.google.com/")

        assert "You're already logged in." in ex.value.stdout.decode("utf-8")


def test_repeated_login_callback_nordvpn_scheme_url():
    """Manual TC: LVPN-460"""

    login.login_as("default")

    with lib.Defer(lambda: sh.nordvpn.logout("--persist-token")):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.login("--callback", "nordvpn://")

        assert "You're already logged in." in ex.value.stdout.decode("utf-8")

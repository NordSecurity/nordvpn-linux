import pytest
import sh

from lib.shell import sh_no_tty

pytestmark = pytest.mark.usefixtures(
    "nordvpnd_scope_module", "unblock_network", "collect_logs"
)


def test_account_output_contains_all_fields():
    """
    Verify that account command output contains all expected fields.

    This test ensures that the account command displays all required fields
    without checking their specific values.
    """
    output = sh_no_tty.nordvpn.account()

    # Header
    assert "Account information" in output

    # Required fields (always present)
    assert "Email address:" in output
    assert "Account created:" in output
    assert "Subscription:" in output
    assert "Dedicated IP:" in output
    assert "Multi-factor authentication (MFA):" in output


def test_account_not_logged_in():
    """Verify account command shows error when not logged in."""
    sh_no_tty.nordvpn.logout("--persist-token", _ok_code=[0, 1])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh_no_tty.nordvpn.account()

    assert "You are not logged in." in ex.value.stdout.decode("utf-8")

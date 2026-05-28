import pytest
import sh

from lib import (
    allowlist,
    firewall,
)

pytestmark = pytest.mark.usefixtures("add_and_delete_random_route", "nordvpnd_scope_function")

# 198.18.0.0/15 is an IETF benchmarking range (RFC 2544) — routable, non-private,
# not subject to LAN discovery rules, safe to use as allowlist test subnets.
NARROWER_A = "198.18.0.0/16"
NARROWER_B = "198.19.0.0/16"
WIDER = "198.18.0.0/15"
TOO_WIDE = "0.0.0.0/1"
NON_OVERLAPPING_A = "198.18.0.0/16"
NON_OVERLAPPING_B = "203.0.113.0/24"


def test_add_wider_subnet_removes_narrower_disconnected():
    allowlist.add_subnet_to_allowlist([NARROWER_A])

    allowlist.add_wider_subnet_to_allowlist(WIDER, expected_removed=[NARROWER_A])

    assert not firewall.is_active(None, [WIDER])


def test_add_wider_subnet_removes_multiple_narrower_disconnected():
    allowlist.add_subnet_to_allowlist([NARROWER_A, NARROWER_B])

    allowlist.add_wider_subnet_to_allowlist(WIDER, expected_removed=[NARROWER_A, NARROWER_B])

    assert not firewall.is_active(None, [WIDER])


def test_add_wider_subnet_declined_leaves_state_unchanged():
    allowlist.add_subnet_to_allowlist([NARROWER_A])

    # Decline the confirmation prompt
    sh.nordvpn(allowlist.get_alias(), "add", "subnet", WIDER, _in="n\n", _ok_code=(0,))

    settings_output = str(sh.nordvpn.settings())
    assert settings_output.count(NARROWER_A) == 1
    assert settings_output.count(WIDER) == 0


def test_add_narrower_subnet_when_wider_exists_rejected():
    # Add the wider subnet first (triggers too-wide warning but succeeds)
    allowlist.add_subnet_to_allowlist([WIDER])

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(allowlist.get_alias(), "add", "subnet", NARROWER_A)

    expected_message = allowlist.MSG_ALLOWLIST_SUBNET_ADD_ERROR % NARROWER_A
    assert expected_message in ex.value.stdout.decode("utf-8")

    settings_output = str(sh.nordvpn.settings())
    assert settings_output.count(WIDER) == 1
    assert settings_output.count(NARROWER_A) == 0


def test_add_subnet_too_wide_shows_warning_and_succeeds():
    cmd_message = sh.nordvpn(allowlist.get_alias(), "add", "subnet", TOO_WIDE)

    assert allowlist.MSG_ALLOWLIST_SUBNET_TOO_WIDE_WARNING in cmd_message
    assert allowlist.MSG_ALLOWLIST_SUBNET_ADD_SUCCESS % TOO_WIDE in cmd_message

    assert str(sh.nordvpn.settings()).count(TOO_WIDE) == 1


def test_add_non_overlapping_subnets_no_prompt():
    allowlist.add_subnet_to_allowlist([NON_OVERLAPPING_A])
    allowlist.add_subnet_to_allowlist([NON_OVERLAPPING_B])

    settings_output = str(sh.nordvpn.settings())
    assert settings_output.count(NON_OVERLAPPING_A) == 1
    assert settings_output.count(NON_OVERLAPPING_B) == 1

import os

import pytest
import sh

import lib
from lib import meshnet, ssh

ssh_client = ssh.Ssh("qa-peer", "root", "root")


def setup_module(module):  # noqa: ARG001
    meshnet.TestUtils.setup_module(ssh_client)


def teardown_module(module):  # noqa: ARG001
    meshnet.TestUtils.teardown_module(ssh_client)


def setup_function(function):  # noqa: ARG001
    meshnet.TestUtils.setup_function(ssh_client)


def teardown_function(function):  # noqa: ARG001
    meshnet.TestUtils.teardown_function(ssh_client)


def test_invite_send():

    assert "Meshnet invitation to 'test@test.com' was sent." in meshnet.send_meshnet_invite("test@test.com")

    assert "test@test.com" in sh.nordvpn.meshnet.invite.list()


def test_invite_send_repeated():
    with lib.Defer(lambda: sh.nordvpn.meshnet.invite.revoke("test@test.com")):
        meshnet.send_meshnet_invite("test@test.com")

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            meshnet.send_meshnet_invite("test@test.com")

        assert "Meshnet invitation for 'test@test.com' already exists." in str(ex.value)


def test_invite_send_own_email():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite(os.environ.get("DEFAULT_LOGIN_USERNAME"))

    assert "Email should belong to a different user." in str(ex.value)


def test_invite_send_not_an_email():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite("test")

    assert "Invalid email 'test'." in str(ex.value)


@pytest.mark.skip(reason="A different error message is expected - LVPN-262")
def test_invite_send_long_email():
    # A long email address containing more than 256 characters is created
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite(email)

    assert "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support." not in str(ex.value)


@pytest.mark.skip(reason="A different error message is expected - LVPN-262")
def test_invite_send_email_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        meshnet.send_meshnet_invite("\u2222@test.com")

    assert "It's not you, it's us. We're having trouble with our servers. If the issue persists, please contact our customer support." not in str(ex.value)


def test_invite_revoke():

    meshnet.send_meshnet_invite("test@test.com")

    assert "Meshnet invitation to 'test@test.com' was revoked." in sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "test@test.com" not in sh.nordvpn.meshnet.invite.list()


def test_invite_revoke_repeated():
    with lib.Defer(lambda: sh.nordvpn.meshnet.invite.revoke("test@test.com")):
        meshnet.send_meshnet_invite("test@test.com")

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_revoke_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_revoke_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_revoke_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.revoke("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)


def test_invite_deny():

    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)

    meshnet.send_meshnet_invite(os.environ.get("QA_PEER_USERNAME"))

    assert os.environ.get("DEFAULT_LOGIN_USERNAME") in ssh_client.exec_command("nordvpn meshnet invite list")
    assert f"Meshnet invitation from '{os.environ.get('DEFAULT_LOGIN_USERNAME')}' was denied." in meshnet.deny_meshnet_invite(ssh_client)


def test_invite_deny_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_deny_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_deny_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.deny("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)


def test_invite_accept():

    meshnet.remove_all_peers()
    meshnet.remove_all_peers_in_peer(ssh_client)
    meshnet.revoke_all_invites()
    meshnet.revoke_all_invites_in_peer(ssh_client)

    meshnet.send_meshnet_invite(os.environ.get("QA_PEER_USERNAME"))

    assert os.environ.get("DEFAULT_LOGIN_USERNAME") in ssh_client.exec_command("nordvpn meshnet invite list")
    assert f"Meshnet invitation from '{os.environ.get('DEFAULT_LOGIN_USERNAME')}' was accepted." in meshnet.accept_meshnet_invite(ssh_client)


def test_invite_accept_non_existent():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept("test@test.com")

    assert "No invitation from 'test@test.com' was found." in str(ex.value)


def test_invite_accept_non_existent_long_email():
    email = "test" * 65 + "@test.com"

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept(email)

    assert f"No invitation from '{email}' was found." in str(ex.value)


def test_invite_accept_non_existent_special_character():
    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.meshnet.invite.accept("\u2222@test.com")

    assert "No invitation from '\u2222@test.com' was found." in str(ex.value)

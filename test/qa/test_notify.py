import pytest
import sh

import lib
from lib import notify, settings


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_module", "collect_logs", "disable_notifications")


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notifications_disabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.is_notify_enabled()

    connect_notification = notify.connect_and_capture_notifications(tech, proto, obfuscated)

    assert connect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)

    disconnect_notification = notify.disconnect_and_capture_notifications()

    assert disconnect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notifications_enabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.is_notify_enabled()

    connect_notification = notify.connect_and_capture_notifications(tech, proto, obfuscated)

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    assert connect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_DETECTED)

    disconnect_notification = notify.disconnect_and_capture_notifications()

    assert disconnect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_DETECTED)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notifications_enabled_connected_disable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.is_notify_enabled()

    connect_notification = notify.connect_and_capture_notifications(tech, proto, obfuscated)

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    assert connect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_DETECTED)

    sh.nordvpn.set.notify.off()
    assert not settings.is_notify_enabled()

    disconnect_notification = notify.disconnect_and_capture_notifications()
    assert disconnect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_NOT_DETECTED)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notifications_disabled_connected_enable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.is_notify_enabled()

    connect_notification = notify.connect_and_capture_notifications(tech, proto, obfuscated)

    assert connect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)

    sh.nordvpn.set.notify.on()
    assert settings.is_notify_enabled()

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    disconnect_notification = notify.disconnect_and_capture_notifications()
    assert disconnect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_DETECTED)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notify_already_enabled_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.is_notify_enabled()

    output = sh.nordvpn.set.notify.on()
    assert notify.NOTIFY_MSG_ERROR_ALREADY_ENABLED in str(output)
    assert settings.is_notify_enabled()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notify_already_enabled_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        sh.nordvpn.set.notify.on()
        assert settings.is_notify_enabled()

        output = sh.nordvpn.set.notify.on()
        assert notify.NOTIFY_MSG_ERROR_ALREADY_ENABLED in str(output)
        assert settings.is_notify_enabled()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notify_already_disabled_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.is_notify_enabled()

    output = sh.nordvpn.set.notify.off()
    assert notify.NOTIFY_MSG_ERROR_ALREADY_DISABLED in str(output)
    assert not settings.is_notify_enabled()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_notify_already_disabled_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not settings.is_notify_enabled()

        output = sh.nordvpn.set.notify.off()
        assert notify.NOTIFY_MSG_ERROR_ALREADY_DISABLED in str(output)
        assert not settings.is_notify_enabled()

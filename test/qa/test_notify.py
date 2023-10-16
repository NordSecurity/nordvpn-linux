from lib import (
    daemon,
    info,
    logging,
    login,
    notify,
    settings
)
import lib
import pytest
import sh
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()

    # Make sure that Notifications are disabled before we execute each test
    lib.set_notify("off")


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notifications_disabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.get_is_notify_enabled()

    connect_notification = notify.connect_base(tech, proto, obfuscated)

    assert connect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)

    disconnect_notification = notify.disconnect_base()

    assert disconnect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notifications_enabled_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.get_is_notify_enabled()
    
    connect_notification = notify.connect_base(tech, proto, obfuscated)

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    assert connect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_DETECTED)

    disconnect_notification = notify.disconnect_base()

    assert disconnect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_DETECTED)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notifications_enabled_connected_disable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.get_is_notify_enabled()

    connect_notification = notify.connect_base(tech, proto, obfuscated)

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    assert connect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_DETECTED)

    sh.nordvpn.set.notify.off()
    assert not settings.get_is_notify_enabled()

    disconnect_notification = notify.disconnect_base()
    assert disconnect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_NOT_DETECTED)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notifications_disabled_connected_enable(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.get_is_notify_enabled()

    connect_notification = notify.connect_base(tech, proto, obfuscated)

    assert connect_notification == notify.NOTIFICATION_NOT_DETECTED, \
        notify.print_tidy_exception(connect_notification, notify.NOTIFICATION_NOT_DETECTED)

    sh.nordvpn.set.notify.on()
    assert settings.get_is_notify_enabled()

    # Should fail here, if tested with 3.16.6, since notification icon is missing
    disconnect_notification = notify.disconnect_base()
    assert disconnect_notification == notify.NOTIFICATION_DETECTED, \
        notify.print_tidy_exception(disconnect_notification, notify.NOTIFICATION_DETECTED)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notify_already_enabled_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.notify.on()
    assert settings.get_is_notify_enabled()

    output = sh.nordvpn.set.notify.on()
    assert notify.NOTIFY_MSG_ERROR_ALREADY_ENABLED in str(output)
    assert settings.get_is_notify_enabled()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notify_already_enabled_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        sh.nordvpn.set.notify.on()
        assert settings.get_is_notify_enabled()

        output = sh.nordvpn.set.notify.on()
        assert notify.NOTIFY_MSG_ERROR_ALREADY_ENABLED in str(output)
        assert settings.get_is_notify_enabled()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notify_already_disabled_disconnected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert not settings.get_is_notify_enabled()

    output = sh.nordvpn.set.notify.off()
    assert notify.NOTIFY_MSG_ERROR_ALREADY_DISABLED in str(output)
    assert not settings.get_is_notify_enabled()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_notify_already_disabled_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with lib.Defer(sh.nordvpn.disconnect):
        sh.nordvpn.connect()

        assert not settings.get_is_notify_enabled()

        output = sh.nordvpn.set.notify.off()
        assert notify.NOTIFY_MSG_ERROR_ALREADY_DISABLED in str(output)
        assert not settings.get_is_notify_enabled()
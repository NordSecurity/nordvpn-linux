import subprocess
from threading import Thread

import sh

from . import server


class NotificationCaptureThreadResult:
    def __init__(self, icon_match: bool, summary_match: bool, body_match: bool):
        self.icon_match = icon_match
        self.summary_match = summary_match
        self.body_match = body_match

    def __eq__(self, other):
        if isinstance(other, NotificationCaptureThreadResult):
            return (self.icon_match == other.icon_match) and (self.summary_match == other.summary_match) and (self.body_match == other.body_match)
        return False


# Used for asserts in tests, [Icon match, Summary match, Body match]
NOTIFICATION_DETECTED = NotificationCaptureThreadResult(True, True, True)
NOTIFICATION_NOT_DETECTED = NotificationCaptureThreadResult(False, False, False)

# Used to check if error messages are correct
NOTIFY_MSG_ERROR_ALREADY_ENABLED = "Notifications are already set to 'enabled'."
NOTIFY_MSG_ERROR_ALREADY_DISABLED = "Notifications are already set to 'disabled'."


def capture_notifications(message):
    """Returns `NotificationCaptureThreadResult`, and contains booleans - icon_match, summary_match, body_match - according to found notification contents."""

    # Timeout is needed, in order for Thread not to hang, as we need to exit the process at some point
    # Timeout can be altered according to how fast you connect to NordVPN server
    command = ["timeout", "6", "dbus-monitor", "--session", "type=method_call,interface=org.freedesktop.Notifications"]
    process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    result = NotificationCaptureThreadResult(False, False, False)

    for line in process.stdout:
        if "/usr/share/icons/hicolor/scalable/apps/nordvpn.svg" in line:
            result.icon_match = True

        if "NordVPN" in line:
            result.summary_match = True

        if message in line:
            result.body_match = True

    return result


class NotificationCaptureThread(Thread):
    def __init__(self, expected_msg):
        Thread.__init__(self)
        self.value: NotificationCaptureThreadResult = NotificationCaptureThreadResult(False, False, False)
        self.expected_message = expected_msg

    def run(self):
        self.value = capture_notifications(self.expected_message)


def connect_and_capture_notifications(tech, proto, obfuscated) -> NotificationCaptureThreadResult:
    """Returns [True, True, True] if notification with all expected contents from NordVPN was captured while connecting to VPN server."""

    # Choose server for test, so we know the full expected message
    name, hostname = server.get_hostname_by(tech, proto, obfuscated)
    expected_msg = f"You are connected to {name} ({hostname})!"

    # We try to capture notifications using other thread when connecting to NordVPN server
    t_connect = NotificationCaptureThread(expected_msg)
    t_connect.start()

    sh.nordvpn.connect(hostname.split(".")[0])

    t_connect.join()

    return t_connect.value
    # Return types, reikia koki structa pakurt ir returnint


def disconnect_and_capture_notifications() -> NotificationCaptureThreadResult:
    """Returns [True, True, True] if notification with all expected contents from NordVPN was captured while disconnecting from VPN server."""

    # We know what message we expect to appear in notification
    expected_msg = "You are disconnected from NordVPN."

    # We try to capture notifications using other thread when disconnecting from NordVPN server
    t_disconnect = NotificationCaptureThread(expected_msg)
    t_disconnect.start()

    sh.nordvpn.disconnect()

    t_disconnect.join()

    return t_disconnect.value


def print_tidy_exception(obj1: NotificationCaptureThreadResult, obj2: NotificationCaptureThreadResult) -> str:
    """Prints values of attributes from specified NotificationCaptureThreadResult type of objects."""
    return \
        "\n\n(icon, summary, body)\n" + \
        f"({obj1.icon_match}, {obj1.summary_match}, {obj1.body_match}) - connect_notification / disconnect_notification - found\n" + \
        f"({obj2.icon_match}, {obj2.summary_match}, {obj2.body_match}) - notify.NOTIFICATION_DETECTED / notify.NOTIFICATION_NOT_DETECTED - expected" + \
        "\n\n"

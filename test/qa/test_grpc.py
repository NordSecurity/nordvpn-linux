import pytest
import threading
import sh
import grpc
from collections.abc import Sequence
from lib import daemon, info, logging, login
from lib.protobuf.daemon import (common_pb2, service_pb2_grpc, state_pb2, status_pb2)

NORDVPND_SOCKET = 'unix:///run/nordvpn/nordvpnd.sock'


def setup_function():  # noqa: ARG001
    daemon.start()
    login.login_as("default")
    logging.log()


def teardown_function():  # noqa: ARG001
    logging.log(data=info.collect())
    logging.log()

    sh.nordvpn.logout("--persist-token")
    sh.nordvpn.set.defaults()
    daemon.stop()


def test_multiple_state_subscribers():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED,
    ]

    num_threads = 5
    threads = []
    results = {}

    threads = [threading.Thread(target=lambda i=i: results.update(
        {i: collect_state_changes(len(expected_states), ['connection_status'])})) for i in range(num_threads)]

    [thread.start() for thread in threads]
    sh.nordvpn.connect()
    [thread.join() for thread in threads]

    for i in range(num_threads):
        assert all(a.connection_status.state == b for a, b in zip(
            results[i], expected_states, strict=True))


@pytest.mark.xfail(reason="LVPN-7891")
def test_tunnel_update_notifications_before_and_after_connect():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED,
        status_pb2.ConnectionState.DISCONNECTED,
    ]

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'])))
    thread.start()
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()
    thread.join()
    assert all(a.connection_status.state == b for a,
               b in zip(result, expected_states, strict=True))


def collect_state_changes(stop_at: int, tracked_states: Sequence[str], timeout: int = 10) -> Sequence[state_pb2.AppState]:
    with grpc.insecure_channel(NORDVPND_SOCKET) as channel:
        stub = service_pb2_grpc.DaemonStub(channel)
        response_stream = stub.SubscribeToStateChanges(
            common_pb2.Empty(), timeout=timeout)
        result = []
        for change in response_stream:
            # Ignore the rest of updates as some settings updates may be published
            if change.WhichOneof('state') in tracked_states:
                result.append(change)
                if len(result) >= stop_at:
                    break
        return result


def test_is_virtual_location_is_true_for_virtual_location():
    check_is_virtual_location_in_response("Algiers", True)


def check_is_virtual_location_in_response(loc: str, expected_is_virtual: bool):
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED
    ]

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'])))
    thread.start()
    sh.nordvpn.connect(loc)
    sh.nordvpn.disconnect()
    thread.join()
    assert result.pop().connection_status.virtualLocation == expected_is_virtual


def test_is_virtual_is_false_for_non_virtual_location():
    check_is_virtual_location_in_response("Poland", False)


def test_manual_connection_source_is_present_in_response():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED
    ]

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'])))
    thread.start()
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()
    thread.join()
    assert result.pop().connection_status.parameters.source == status_pb2.ConnectionSource.MANUAL

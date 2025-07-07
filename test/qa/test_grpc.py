import pytest
import threading
import sh
import grpc
from collections.abc import Sequence
from lib.protobuf.daemon import (common_pb2, service_pb2_grpc, state_pb2, status_pb2)
from threading import Barrier


pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


NORDVPND_SOCKET = 'unix:///run/nordvpn/nordvpnd.sock'


def test_multiple_state_subscribers():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED,
    ]

    num_threads = 5
    all_threads_subscribed = Barrier(num_threads+1) # number of child threads + parrent thread
    threads = []
    results = {}

    with grpc.insecure_channel(NORDVPND_SOCKET) as channel:
        threads = [threading.Thread(target=lambda i=i: results.update(
            {i: collect_state_changes(channel, len(expected_states), ['connection_status'], all_threads_subscribed)})) for i in range(num_threads)]

        [thread.start() for thread in threads]
        all_threads_subscribed.wait(timeout=10)
        sh.nordvpn.connect()
        [thread.join() for thread in threads]

        for i in range(num_threads):
            assert all(a.connection_status.state == b for a, b in zip(
                results[i], expected_states, strict=True))


def test_tunnel_update_notifications_before_and_after_connect():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED,
        status_pb2.ConnectionState.DISCONNECTED,
    ]

    subscribtion_barrier = Barrier(2) # parrent and child thread
    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes_guard(
        len(expected_states), ['connection_status'], subscribtion_barrier)))
    thread.start()
    subscribtion_barrier.wait(timeout=10)
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()
    thread.join()
    assert all(a.connection_status.state == b for a,
               b in zip(result, expected_states, strict=True))


def collect_state_changes_guard(stop_at: int, tracked_states: Sequence[str], subscribtion_barrier: Barrier, timeout: int = 30) -> Sequence[state_pb2.AppState]:
    with grpc.insecure_channel(NORDVPND_SOCKET) as channel:
        return collect_state_changes(channel, stop_at, tracked_states, subscribtion_barrier, timeout)

def collect_state_changes(channel: grpc.Channel, stop_at: int, tracked_states: Sequence[str], subscribtion_barrier: Barrier, timeout: int = 30) -> Sequence[state_pb2.AppState]:
        grpc.channel_ready_future(channel).result(timeout=timeout)
        stub = service_pb2_grpc.DaemonStub(channel)
        subscribtion_barrier.wait()
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

    subscribtion_barrier = Barrier(2) # parrent and child thread
    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes_guard(
        len(expected_states), ['connection_status'], subscribtion_barrier)))
    thread.start()
    subscribtion_barrier.wait(timeout=10)
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

    subscribtion_barrier = Barrier(2) # parrent and child thread
    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes_guard(
        len(expected_states), ['connection_status'], subscribtion_barrier)))
    thread.start()
    subscribtion_barrier.wait(timeout=10)
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()
    thread.join()
    assert result.pop().connection_status.parameters.source == status_pb2.ConnectionSource.MANUAL

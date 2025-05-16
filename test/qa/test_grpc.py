import threading
import sh
import grpc
import time
import traceback
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
    results = {}
    exceptions = {}
    lock = threading.Lock()
    barrier = threading.Barrier(num_threads + 1)  # +1 for the main thread

    def state_subscriber_worker(i):
        try:
            barrier.wait()
            states = collect_state_changes(len(expected_states), ['connection_status'])
            with lock:
                results[i] = states
        except BaseException: # noqa: BLE001
            with lock:
                exceptions[i] = traceback.format_exc()

    threads = [
        threading.Thread(target=state_subscriber_worker, args=(i,)) for i in range(num_threads)
    ]

    for thread in threads:
        thread.start()

    # make sure threads started
    barrier.wait()

    sh.nordvpn.connect()

    for thread in threads:
        thread.join()

    if exceptions:
        raise RuntimeError("Exceptions in threads:\n" + "\n".join(exceptions.values()))

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

    result = []
    barrier = threading.Barrier(2)  # One for main thread, one for worker

    thread = threading.Thread(
        target=state_listener_worker,
        args=(expected_states, result, barrier)
    )
    thread.start()

    barrier.wait()

    sh.nordvpn.connect()
    time.sleep(5)
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
    barrier = threading.Barrier(2)  # One for main thread, one for worker

    thread = threading.Thread(
        target=state_listener_worker,
        args=(expected_states, result, barrier)
    )
    thread.start()

    barrier.wait()

    sh.nordvpn.connect(loc)
    time.sleep(5)
    sh.nordvpn.disconnect()

    thread.join()

    assert result.pop().connection_status.virtualLocation == expected_is_virtual


def test_is_virtual_is_false_for_non_virtual_location():
    check_is_virtual_location_in_response("Poland", False)


def state_listener_worker(expected_states, result_container, barrier):
    barrier.wait()
    result_container.extend(collect_state_changes(len(expected_states), ['connection_status']))


def test_manual_connection_source_is_present_in_response():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED
    ]

    result = []
    barrier = threading.Barrier(2)  # One for the main thread, one for the worker

    thread = threading.Thread(
        target=state_listener_worker,
        args=(expected_states, result, barrier)
    )
    thread.start()

    barrier.wait()

    sh.nordvpn.connect()
    time.sleep(5)
    sh.nordvpn.disconnect()

    thread.join()

    assert result.pop().connection_status.parameters.source == status_pb2.ConnectionSource.MANUAL

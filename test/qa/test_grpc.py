import pytest
import threading
import sh
import grpc
import datetime
from collections.abc import Sequence
from lib.protobuf.daemon import (common_pb2, service_pb2_grpc, state_pb2, status_pb2)
from lib import logging
import json

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


NORDVPND_SOCKET = 'unix:///run/nordvpn/nordvpnd.sock'


def test_multiple_state_subscribers():
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED,
    ]

    num_threads = 5
    sem = threading.Barrier(num_threads + 1)
    threads = []
    results = {}

    chan = grpc.insecure_channel(NORDVPND_SOCKET)

    def collect_state_changes_single_chan(stop_at: int, tracked_states: Sequence[str], subscribed_semaphore: threading.Barrier, channel: grpc.Channel, timeout: int = 15) -> Sequence[state_pb2.AppState]:
        grpc.channel_ready_future(channel).result(timeout=timeout)
        logging.log(f"DEBUG: subscribe to state changes: {datetime.datetime.now()}")
        stub = service_pb2_grpc.DaemonStub(channel)
        response_stream = stub.SubscribeToStateChanges(
            common_pb2.Empty(), timeout=timeout)
        subscribed_semaphore.wait()
        logging.log(f"DEBUG: subscribed: {datetime.datetime.now()}")
        result = []
        for change in response_stream:
            logging.log(f"DEBUG: received state change: {change}")
            # Ignore the rest of updates as some settings updates may be published
            if change.WhichOneof('state') in tracked_states:
                result.append(change)
                if len(result) >= stop_at:
                    break
        logging.log(f"DEBUG: state changes collected: {datetime.datetime.now()}")
        response_stream.cancel()
        return result

    threads = [threading.Thread(target=lambda i=i: results.update(
        {i: collect_state_changes_single_chan(len(expected_states), ['connection_status'], sem, chan)})) for i in range(num_threads)]

    [thread.start() for thread in threads]
    sem.wait()
    sh.nordvpn.connect()
    [thread.join() for thread in threads]

    chan.close()

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

    sem = threading.Barrier(2)

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'], sem)))
    thread.start()
    sem.wait()
    logging.log(f"DEBUG: connect: {datetime.datetime.now()}")
    sh.nordvpn.connect()
    logging.log(f"DEBUG: connected: {datetime.datetime.now()}")
    sh.nordvpn.disconnect()
    logging.log(f"DEBUG: disconnected: {datetime.datetime.now()}")
    thread.join()
    assert all(a.connection_status.state == b for a,
               b in zip(result, expected_states, strict=True))


def collect_state_changes(stop_at: int, tracked_states: Sequence[str], subscribed_semaphore: threading.Barrier, timeout: int = 15) -> Sequence[state_pb2.AppState]:
    logging.log(f"DEBUG: subscribe to state changes: {datetime.datetime.now()}")
    with grpc.insecure_channel(NORDVPND_SOCKET) as channel:
        grpc.channel_ready_future(channel).result(timeout=timeout)
        stub = service_pb2_grpc.DaemonStub(channel)
        response_stream = stub.SubscribeToStateChanges(
            common_pb2.Empty(), timeout=timeout)
        subscribed_semaphore.wait()
        logging.log(f"DEBUG: subscribed: {datetime.datetime.now()}")
        result = []
        for change in response_stream:
            logging.log(f"DEBUG: received state change: {change}")
            # Ignore the rest of updates as some settings updates may be published
            if change.WhichOneof('state') in tracked_states:
                result.append(change)
                if len(result) >= stop_at:
                    break
        logging.log(f"DEBUG: state changes collected: {datetime.datetime.now()}")
        response_stream.cancel()
        channel.close()
        return result


def test_is_virtual_location_is_true_for_virtual_location():
    check_is_virtual_location_in_response("Algiers", True)


def check_is_virtual_location_in_response(loc: str, expected_is_virtual: bool):
    expected_states = [
        status_pb2.ConnectionState.CONNECTING, # start with "connecting" state ASAP
        status_pb2.ConnectionState.CONNECTING, # update with selected location
        status_pb2.ConnectionState.CONNECTED
    ]

    sem = threading.Barrier(2)

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'], sem)))
    thread.start()
    sem.wait()
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

    sem = threading.Barrier(2)

    result = []
    thread = threading.Thread(target=lambda: result.extend(collect_state_changes(
        len(expected_states), ['connection_status'], sem)))
    thread.start()
    sem.wait()
    sh.nordvpn.connect()
    sh.nordvpn.disconnect()
    thread.join()
    assert result.pop().connection_status.parameters.source == status_pb2.ConnectionSource.MANUAL

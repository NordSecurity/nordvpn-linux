from lib import (
    daemon,
    info,
    logging,
    login,
    network,
    server,
    settings
)
import lib
import pytest
import queue
import sh
import socket
import threading
import time
import timeout_decorator


def setup_module(module):
    daemon.start()
    login.login_as("default")


def teardown_module(module):
    sh.nordvpn.logout("--persist-token")
    daemon.stop()


def setup_function(function):
    logging.log()


def teardown_function(function):
    logging.log(data=info.collect())
    logging.log()


def capture_traffic() -> int:
    """
    Captures traffic that goes to VPN server 
    :return: int - returns count of captured packets
    """

    # Collect information needed for tshark filter
    server_ip = settings.get_server_ip()
    protocol = settings.get_current_connection_protocol()
    obfuscated = settings.get_is_obfuscated()

    # Choose traffic filter according to information collected above
    if protocol == "nordlynx":
        traffic_filter = "(udp port 51820) and (ip dst {})".format(server_ip)
    elif protocol == "udp" and not obfuscated:
        traffic_filter = "(udp port 1194) and (ip dst {})".format(server_ip)
    elif protocol == "tcp" and not obfuscated:
        traffic_filter = "(tcp port 443) and (ip dst {})".format(server_ip)
    elif protocol == "udp" and obfuscated:
        traffic_filter = "udp and (port not 1194) and (ip dst {})".format(server_ip)
    elif protocol == "tcp" and obfuscated:
        traffic_filter = "tcp and (port not 443) and (ip dst {})".format(server_ip)

    # Actual capture
    # If 2 packets were already captured, do not wait for 3 seconds
    # Show compact output about packets
    tshark_result = sh.tshark("-i", "any", "-T", "fields", "-e", "ip.src", "-e", "ip.dst", "-a", "duration:3", "-a", "packets:2", "-f", traffic_filter)

    packets = tshark_result.replace("\t", " -> ")
    packets = tshark_result.split("\n")

    logging.log("PACKETS_CAPTURED: " + str(packets))

    # If no packets were captured, `packets` value should be 0
    return len(packets) - 1


def connect_base_test(group=[], name="", hostname=""):
    output = sh.nordvpn.connect(group)
    print(output)

    # Start capturing packets
    packet_capture_thread_queue = queue.Queue()
    packet_capture_thread_lambda = lambda: packet_capture_thread_queue.put(capture_traffic())
    packet_capture_thread = threading.Thread(target=packet_capture_thread_lambda)
    packet_capture_thread.start()

    # We need to make sure, that packets are being sent out only after
    # tshark starts, and not earlier, so we wait for one second.
    time.sleep(1)
    assert lib.is_connect_successful(output, name, hostname)

    # Following function creates atleast two ICMP packets
    assert network.is_connected()

    packet_capture_thread.join()
    packet_capture_thread_result = packet_capture_thread_queue.get()

    assert packet_capture_thread_result >= 2


def disconnect_base_test():
    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
    assert "nordlynx" not in sh.ip.a() and "nordtun" not in sh.ip.a()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_quick_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test()
    disconnect_base_test()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_quick_connect_double_only(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for n in range(2):
        connect_base_test()

    disconnect_base_test()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_connect_to_server_absent(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("moon")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)
    assert network.is_disconnected()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
def test_mistype_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.kinect()

    print(ex.value)
    assert lib.is_invalid_command("kinect", ex)
    assert network.is_disconnected()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_server_random_by_name(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    name, hostname = server.get_hostname_by(tech, proto, obfuscated)
    connect_base_test(hostname.split(".")[0], name, hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_random_server_by_name_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    name, hostname = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test(hostname.split(".")[0], name, hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_random_server_by_name_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    name, hostname = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test(hostname.split(".")[0], name, hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_random_server_by_name_ovpn(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    name, hostname = server.get_hostname_by(group_id=group)
    connect_base_test(hostname.split(".")[0], name, hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OBFUSCATED_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_random_server_by_name_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    name, hostname = server.get_hostname_by(group_id=group)
    connect_base_test(hostname.split(".")[0], name, hostname)
    disconnect_base_test()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_network_restart(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test()

    links = socket.if_nameindex()
    logging.log(links)
    default_gateway = network.stop()
    network.start(default_gateway)
    daemon.wait_for_reconnect(links)
    with lib.ErrorDefer(sh.nordvpn.disconnect):
        assert network.is_connected()
    logging.log(info.collect())

    disconnect_base_test()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_quick_connect_double_disconnect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for n in range(2):
        connect_base_test()
        disconnect_base_test()


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_network_gone(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    default_gateway = network.stop()
    with lib.Defer(lambda: network.start(default_gateway)):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.connect()
        print(ex.value)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_ovpn(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OBFUSCATED_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_flag_group_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_flag_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)

@pytest.mark.parametrize("group", lib.OVPN_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OVPN_STANDARD_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_flag_group_ovpn(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.OBFUSCATED_TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_flag_group_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_group_invalid(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn.connect("--group", "nonexisting_group")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("country", lib.COUNTRIES)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_country(tech, proto, obfuscated, country):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(country)
    disconnect_base_test()


@pytest.mark.parametrize("country_code", lib.COUNTRY_CODES)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_country_code(tech, proto, obfuscated, country_code):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(country_code)
    disconnect_base_test()


@pytest.mark.parametrize("city", lib.CITIES)
@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_city(tech, proto, obfuscated, city):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test(city)
    disconnect_base_test()


def get_unavailable_groups():
    """ Returns groups that are not available with current connection settings """
    ALL_GROUPS = ['Africa_The_Middle_East_And_India',
              'Asia_Pacific',
              'Dedicated_IP',
              'Double_VPN',
              'Europe',
              'Obfuscated_Servers',
              'Onion_Over_VPN',
              'P2P',
              'Standard_VPN_Servers',
              'The_Americas']

    CURRENT_GROUPS = str(sh.nordvpn.groups()).strip("%-\r  ").strip().split(", ")

    return set(ALL_GROUPS) - set(CURRENT_GROUPS)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_unavailable_groups(tech, proto, obfuscated):
    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    UNAVAILABLE_GROUPS = get_unavailable_groups()
    logging.log("UNAVAILABLE_GROUPS: " + str(UNAVAILABLE_GROUPS))

    for group in UNAVAILABLE_GROUPS:
        logging.log("CHECKING_GROUP: " + group)
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.connect(group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("tech,proto,obfuscated", lib.TECHNOLOGIES)
@pytest.mark.flaky(reruns=2, reruns_delay=90)
@timeout_decorator.timeout(40)
def test_connect_to_unavailable_servers(tech, proto, obfuscated):
    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    UNAVAILABLE_GROUPS = get_unavailable_groups()
    logging.log("UNAVAILABLE_GROUPS: " + str(UNAVAILABLE_GROUPS))

    for group in UNAVAILABLE_GROUPS:
        name = server.get_hostname_by(group_id=group)[1].split(".")[0]
        logging.log("CHECKING_GROUP: " + group)
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn.connect(name)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)
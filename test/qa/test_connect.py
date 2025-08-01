import random
import socket
import time

import pytest
import sh

import lib
from lib import daemon, info, logging, network, server

pytestmark = pytest.mark.usefixtures("nordvpnd_scope_function")


CONNECT_ALIAS = [
    "connect",
    "c"
]


def get_alias() -> str:
    """
    This function randomly picks an alias from the predefined list 'CONNECT_ALIAS' and returns it.

    Returns:
        str: A randomly selected alias from CONNECT_ALIAS.
    """
    return random.choice(CONNECT_ALIAS)


def connect_base_test(connection_settings, group=(), name="", hostname=""):
    print(connection_settings)
    output = sh.nordvpn(get_alias(), group, _tty_out=False)
    print(output)

    assert lib.is_connect_successful(output, name, hostname)
    assert network.is_available()


def disconnect_base_test():
    output = sh.nordvpn.disconnect()
    print(output)
    assert lib.is_disconnect_successful(output)
    assert network.is_disconnected()
    assert "nordlynx" not in sh.ip.a() and "nordtun" not in sh.ip.a() and "qtun" not in sh.ip.a()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_quick_connect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_server_absent(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "moon")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)
    assert network.is_disconnected()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_server_random_by_name(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_random_server_by_name_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_group_random_server_by_name_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_random_server_by_name_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(tech, proto, obfuscated, group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)

    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_group_random_server_by_name_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_hostname_by(group_name=group)
    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()


@pytest.mark.skip("flaky test, LVPN-6277")
# the tun interface is recreated only for OpenVPN
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_network_restart_recreates_tun_interface(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))

    links = socket.if_nameindex()
    logging.log(links)
    default_gateway = network.stop()
    network.start(default_gateway)
    daemon.wait_for_reconnect(links)
    assert network.is_connected()
    logging.log(info.collect())

    disconnect_base_test()


# for Nordlynx normally the tunnel is not recreated
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_connect_network_restart_nordlynx(tech, proto, obfuscated):
    if daemon.is_init_systemd():
        pytest.skip("LVPN-5733")

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated))

    links = socket.if_nameindex()
    logging.log(links)
    default_gateway = network.stop()
    network.start(default_gateway)

    # wait for internet
    network.is_available(10)

    assert network.is_connected()
    assert links == socket.if_nameindex()

    logging.log(info.collect())

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_quick_connect_double_disconnect(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    for _ in range(2):
        connect_base_test((tech, proto, obfuscated))
        disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_network_gone(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    default_gateway = network.stop()
    with lib.Defer(lambda: network.start(default_gateway)):
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias())
        print(ex.value)


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.DEDICATED_IP_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_group_ovpn(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_group_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), group)
    disconnect_base_test()


@pytest.mark.parametrize("group", lib.STANDARD_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_flag_group_standard(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_flag_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.ADDITIONAL_GROUPS_NORDWHISPER)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.NORDWHISPER_TECHNOLOGY)
def test_nordwhisper_connect_to_flag_group_additional(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.DEDICATED_IP_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_flag_group_ovpn(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("group", lib.OVPN_OBFUSCATED_GROUPS)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.OBFUSCATED_TECHNOLOGIES)
def test_connect_to_flag_group_obfuscated(tech, proto, obfuscated, group):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), ["--group", group])
    disconnect_base_test()

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", group, group)

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_group_invalid(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    with pytest.raises(sh.ErrorReturnCode_1) as ex:
        sh.nordvpn(get_alias(), "--group", "nonexistent_group")

    print(ex.value)
    assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize("country", lib.COUNTRIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_country(tech, proto, obfuscated, country):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), country)
    disconnect_base_test()


@pytest.mark.parametrize("country_code", lib.COUNTRY_CODES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_country_code(tech, proto, obfuscated, country_code):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), country_code)
    disconnect_base_test()


@pytest.mark.parametrize("city", lib.CITIES)
@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_city(tech, proto, obfuscated, city):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    connect_base_test((tech, proto, obfuscated), city)
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_connect_to_unavailable_groups(tech, proto, obfuscated):
    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias(), group)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_connect_to_unavailable_servers(tech, proto, obfuscated):
    # TODO: LVPN-257
    time.sleep(3)

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    unavailable_groups = daemon.get_unavailable_groups()

    for group in unavailable_groups:
        server_info = server.get_hostname_by(group_name=group)
        name = server_info.hostname.split(".")[0]

        with pytest.raises(sh.ErrorReturnCode_1) as ex:
            sh.nordvpn(get_alias(), name)

        print(ex.value)
        assert lib.is_connect_unsuccessful(ex)


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES)
def test_status_connected(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    assert network.is_disconnected()
    assert "Disconnected" in sh.nordvpn.status()

    server_info = server.get_hostname_by(technology=tech, protocol=proto, obfuscated=obfuscated)
    sh.nordvpn(get_alias(), server_info.hostname.split(".")[0])

    connect_time = time.monotonic()

    network.generate_traffic(retry=5)

    status_info = daemon.get_status_data()
    status_time = time.monotonic()

    print("status_info: " + str(status_info))
    print("actual_status: " + str(sh.nordvpn.status()))

    assert "Connected" in status_info["status"]

    assert server_info.hostname in status_info["hostname"]
    assert server_info.name in status_info["server"]

    assert socket.gethostbyname(server_info.hostname) in status_info["ip"]

    assert server_info.country in status_info["country"]
    assert server_info.city in status_info["city"]

    assert tech.upper() in status_info["current technology"]

    if tech == "openvpn":
        assert proto.upper() in status_info["current protocol"]
    elif tech == "nordwhisper":
        assert "Webtunnel" in status_info["current protocol"]
    else:
        assert "UDP" in status_info["current protocol"]

    transfer_received = float(status_info["transfer"].split(" ")[0])
    transfer_sent = float(status_info["transfer"].split(" ")[3])

    assert transfer_received >= 0
    assert transfer_sent > 0

    time_connected = int(status_info["uptime"].split(" ")[0])
    time_passed = status_time - connect_time
    if "minute" in status_info["uptime"]:
        time_connected_seconds = int(status_info["uptime"].split(" ")[2])
        assert time_passed - 1 <= time_connected * 60 + time_connected_seconds <= time_passed + 1
    else:
        assert time_passed - 1 <= time_connected <= time_passed + 1

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES)
def test_connect_to_virtual_server(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)
    sh.nordvpn.set("virtual-location", "on")
    virtual_countries = lib.get_virtual_countries()

    assert len(virtual_countries) > 0
    country = random.choice(virtual_countries)

    connect_base_test((tech, proto, obfuscated), country)
    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.TECHNOLOGIES_BASIC1)
def test_connect_to_post_quantum_server(tech, proto, obfuscated):

    lib.set_technology_and_protocol(tech, proto, obfuscated)

    sh.nordvpn.set.pq("on")

    connect_base_test((tech, proto, obfuscated))

    assert "preshared key" in sh.sudo.wg.show()

    disconnect_base_test()


def test_check_routing_table_for_lan():
    # check that the routing table is correctly configured when LAN is enabled and that the tunnel IP is correct
    lib.set_technology_and_protocol("nordlynx", "", "")

    default_route = network.RouteInfo.default_route_info()
    connect_base_test(("nordlynx", "", ""))

    # check the tunnel IP
    nordlynx_route = network.RouteInfo(sh.ip.route.show.dev("nordlynx").stdout.decode())
    assert nordlynx_route.destination == "10.5.0.0/16"
    assert nordlynx_route.src == "10.5.0.2"

    lan_ips = [
        "10.0.0.1",
        "172.16.0.1",
        "192.168.0.1",
        "169.254.0.1",
    ]

    private_vpn_ip = "10.5.0.1"

    # check LAN IP is not routed thru main
    assert all(not default_route.routes_ip(ip) for ip in lan_ips)
    assert not default_route.routes_ip(private_vpn_ip)
    assert nordlynx_route.routes_ip(private_vpn_ip)

    # enable LAN discovery
    sh.nordvpn.set("lan-discovery", "on")

    # check LAN IP is routed thru main
    assert all(default_route.routes_ip(ip) for ip in lan_ips)

    # IP from VPN private range is routed thru nordlynx interface
    assert not default_route.routes_ip(private_vpn_ip)
    assert nordlynx_route.routes_ip(private_vpn_ip)

    disconnect_base_test()


@pytest.mark.parametrize(("tech", "proto", "obfuscated"), lib.STANDARD_TECHNOLOGIES_NO_NORDWHISPER)
def test_connect_to_dedicated_ip(tech, proto, obfuscated):
    lib.set_technology_and_protocol(tech, proto, obfuscated)

    server_info = server.get_dedicated_ip()

    connect_base_test((tech, proto, obfuscated), server_info.hostname.split(".")[0], server_info.name, server_info.hostname)
    disconnect_base_test()

    assert network.is_disconnected()
    assert "nordlynx" not in sh.ip.a() and "nordtun" not in sh.ip.a()

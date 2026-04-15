import os
import socket
import struct
import time

import sh

from . import Port, Protocol, ssh

IP_ROUTE_TABLE = 205
ENDPOINTS = "endpoints"
SOCK_TIMEOUT = 5
TCP_DST_PORT = 1234
UDP_DST_PORT = 1235
DST_ADDR = "qa-peer"
# This is used when communicating with the socket server
# on the qa-peer container, used in port allowlisting behavioral testing
try:
    IP = socket.gethostbyname(DST_ADDR)
except Exception as e: # noqa: BLE001
    print(e)
    print("Unable to get qa-peer addr, does it exist?")

LAN_DISCOVERY_SUBNETS = [
    "169.254.0.0/16",
    "192.168.0.0/16",
    "172.16.0.0/12",
    "10.0.0.0/8"
]

def setup_port_sock_server(ssh_client : ssh.Ssh | None):
    if ssh_client is None:
        ssh_client = ssh.Ssh("qa-peer", "root", "root")
    ssh_client.connect()
    remote_path = "/tmp/sockserv.py"
    project_root = os.environ["WORKDIR"]
    serv_path = f"{project_root}/test/qa/lib/sockserv.py"
    ssh_client.send_file(serv_path, remote_path)
    command = (
        f"nohup python3 {remote_path}"
        f"> {project_root}/server.log 2>&1 &"
    )
    out = ssh_client.exec_command(command)
    print(out)
    print(f"Remote script started in background: {remote_path}")
    time.sleep(1)


def is_active() -> bool:
    """Returns True when nft finds the default nordvpn table"""
    print(sh.ip.route())
    try:
        out = sh.sudo.nft("list", "table", "inet", "nordvpn")
    except Exception: # noqa: BLE001
        return False

    print(out)
    print(sh.nordvpn.settings())
    return "nordvpn" in out

tun_interface_names = [
    "nordtun",
    "qtun",
    "nordlynx"
]

def is_ip_routed_via_VPN(subnets: list[str]) -> bool:
    bool_array = []
    for subnet in subnets:
        print(sh.ip.route.get(subnet))
        # Allowlisted subnet should not return tunnel interface name when using ip route get
        bool_array.append(any(iface_name in sh.ip.route.get(subnet) for iface_name in tun_interface_names))
    return all(bool_array)



def is_source_port_reachable(ports: list[Port]) -> bool:
    """Must be used in a test in which setup_port_sock_server is setup, otherwise will not work"""
    bool_array = []
    for port in ports:
        # Given port is a range `3000:3100`, in such a case we wish to test both ends of the range
        if ":" in port.value:
            port_range_start, port_range_end = port.value.split(":")
            bool_array.append(process_port(Port(port_range_start, port.protocol)) and process_port(Port(port_range_end, port.protocol)))
        else:
            bool_array.append(process_port(port))
        # clean up connmarks for test connections as it affects next test
        for protocol in ["TCP", "UDP"]:
            try:
                print("deleting ", protocol)
                sh.sudo.conntrack("-D", "-p", protocol, "--sport", port.value)
            except sh.ErrorReturnCode_1:
                print("nothing to delete")
                continue
    return all(bool_array)



def process_port(port: Port) -> bool:
    if port.protocol == Protocol.TCP:
        print("process tcp")
        return is_port_accessible_TCP(int(port.value))
    if port.protocol == Protocol.UDP:
        print("process udp")
        return is_port_accessible_UDP(int(port.value))
    print("process both, tcp")
    tcp_retval = is_port_accessible_TCP(int(port.value))
    print("process both, udp")
    udp_retval = is_port_accessible_UDP(int(port.value))
    return tcp_retval and udp_retval


def is_port_accessible_TCP(src_port : int) -> bool :
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    s.bind(("0.0.0.0", src_port))
    s.setsockopt(
        socket.SOL_SOCKET,
        socket.SO_LINGER,
        struct.pack('ii', 1, 0)
    )
    s.settimeout(SOCK_TIMEOUT)
    try:
        s.connect((IP, TCP_DST_PORT))
        s.send(b"ping")
        print("TCP data sent")
        data = s.recv(4096)
        print("data received: ", data)
        retval = True
    except TimeoutError:
        print(f"timeout of {SOCK_TIMEOUT} hit with TCP, source port {src_port}, dst port : {TCP_DST_PORT}",)
        retval = False
    # `OSError: [Errno 99] Cannot assign requested address` handling
    except OSError as e:
        print(e)
        retval = False
    finally:
        s.close()
    return retval


def is_port_accessible_UDP(src_port):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", src_port))
    s.settimeout(SOCK_TIMEOUT)
    try:
        s.sendto(b"ping", (IP, UDP_DST_PORT))
        print("UDP data sent")
        data, addr = s.recvfrom(4096)
        print("data received back from server: ", data)
        return True
    except PermissionError:
        print("unable to send packet to address")
        return False
    except TimeoutError:
        print(f"timeout of {SOCK_TIMEOUT} hit with UDP, source port {src_port}, dst port : {UDP_DST_PORT}",)
        return False
    finally:
        s.close()


def add_and_delete_random_route():
    """Adds a random route, and deletes it. If this is not used, exceptions happen in allowlist tests."""
    # cmd = sh.sudo.ip.route.add.default.via.bake("127.0.0.1")
    # cmd.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route add default via 127.0.0.1 table {IP_ROUTE_TABLE}").read()
    # sh.sudo.ip.route.delete.default.table(IP_ROUTE_TABLE)
    os.popen(f"sudo ip route delete default table {IP_ROUTE_TABLE}").read()

import socket
# import lib
import argparse

#UDP
def udp_server(port):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", port))
    while True:
        data, addr = s.recvfrom(4096)
        s.sendto(b"pong", addr)
        print(data)


# TCP
def tcp_server(port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(("0.0.0.0", port))
    sock.listen(5)
    while True:
        conn, _ = sock.accept()
        data = conn.recv(4096)
        conn.send(b"pong")
        print(data)


parser = argparse.ArgumentParser(description="Select network protocol")
parser.add_argument(
    "-p", "--protocol",
    choices=["tcp", "udp"],
    default="tcp",
    help="Choose protocol: tcp or udp (default: tcp)",
)
args = parser.parse_args()
if args.protocol == "tcp":
    tcp_server(1234)
else:
    udp_server(1235)
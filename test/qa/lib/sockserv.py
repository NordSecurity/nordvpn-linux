import socket

# import lib
import argparse
import threading


# UDP
def udp_server(port):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", port))
    while True:
        data, addr = s.recvfrom(4096)
        ip, p = addr
        s.sendto(b"pong", addr)
        print("udp", data, ip, p)


# TCP
def tcp_server(port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(("0.0.0.0", port))
    sock.listen(5)
    while True:
        conn, addr = sock.accept()
        data = conn.recv(4096)
        conn.send(b"pong")
        ip, p = addr
        print("tcp", data, ip, p)


parser = argparse.ArgumentParser(description="Select network protocol")
parser.add_argument(
    "-p",
    "--protocol",
    choices=["tcp", "udp"],
    default="tcp",
    help="Choose protocol: tcp or udp (default: tcp)",
)
args = parser.parse_args()

t1 = threading.Thread(target=tcp_server, args=(1234,), daemon=True)
t2 = threading.Thread(target=udp_server, args=(1235,), daemon=True)

t1.start()
t2.start()

t1.join()
t2.join()

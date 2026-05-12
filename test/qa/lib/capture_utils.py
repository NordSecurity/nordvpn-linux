import asyncio
import threading
import os
from collections import defaultdict

import pyshark

class SLL2LiveCapture(pyshark.LiveCapture):
    def _get_dumpcap_parameters(self):
        params = super()._get_dumpcap_parameters()
        params += ["-y", "LINUX_SLL2"]
        return params

class BackgroundCapture:
    def __init__(self, interface, bpf_filter=None, display_filter=None):
        self.interface = interface
        self.bpf_filter = bpf_filter
        self.display_filter = display_filter
        self.packets = []
        self._thread = None
        self._capture = None
        self._loop = None
        self._ready = threading.Event()

    def _run(self):
        # Each thread needs its own asyncio loop for pyshark
        self._loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self._loop)

        self._capture = SLL2LiveCapture(
            interface=self.interface,
            bpf_filter=self.bpf_filter,
            display_filter=self.display_filter,
        )

        try:
            for pkt in self._capture.sniff_continuously():
                self.packets.append(pkt)
        except Exception as e: # noqa: BLE001
            print(f"[capture] stopped: {e}")

    def start(self):
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()


    def stop(self):
        try:
            if self._capture is not None:
                self._capture.close()
        except Exception: # noqa: BLE001
            pass
        if self._thread is not None:
            self._thread.join(timeout=3)
        self.packets = sorted(self.packets, key=lambda p: p.sniff_time)

def ifindex_to_name(idx, _cache=None):
    if _cache is None:
        _cache={}
    idx = int(idx)
    if idx in _cache:
        return _cache[idx]
    _cache.clear()
    for name in os.listdir("/sys/class/net"):
        try:
            with open(f"/sys/class/net/{name}/ifindex") as f:
                _cache[int(f.read())] = name
        except OSError:
            pass
    return _cache.get(idx, f"if{idx}")

# helper func for debugging, not used in tests themselves
def summarize(packets):
    print(f"\nCaptured {len(packets)} packets")

    for i, pkt in enumerate(packets, 1):
        try:
            proto = pkt.highest_layer
            ifindex = ifindex_to_name(pkt.sll.ifindex)
            src = pkt.ip.src if hasattr(pkt, "ip") else "?"
            dst = pkt.ip.dst if hasattr(pkt, "ip") else "?"
            length = pkt.length
            timestamp = pkt.sniff_time
            print(f"{i:3d}. {timestamp} {ifindex} {proto:8s} {src} -> {dst}  len={length}")
        except AttributeError:
            pass

def check_for_routing_pattern(packets: list, peer_ip, endpoint_ip):
    groups = defaultdict(list)
    for pkt in packets:
        groups[pkt.length].append(pkt)

    grouped_packets = list(groups.values())
    print(grouped_packets)
    handshake_group = grouped_packets[0]
    for index, _ in enumerate(handshake_group):
        try:
            # initial request of peer -> endpoint
            packet1 = handshake_group[index]
            packet2 = handshake_group[index + 1]
            packet3 = handshake_group[index + 2]
            packet4 = handshake_group[index + 3]
            # initial request of peer -> endpoint
            if (ifindex_to_name(packet1.sll.ifindex) == "nordlynx" and packet1.ip.src == peer_ip and packet1.ip.dst == endpoint_ip and
            # second packet in order is the device being routed through's interface passing to endpoint
                ifindex_to_name(packet2.sll.ifindex) != "nordlynx" and packet2.ip.dst == endpoint_ip and
            # third packet is endpoint ip sending ack to router
                ifindex_to_name(packet3.sll.ifindex) != "nordlynx" and packet3.ip.src == endpoint_ip and
            # fourth packet is sending this ack now back to the peer over the tunnel
                ifindex_to_name(packet4.sll.ifindex) == "nordlynx" and packet4.ip.src == endpoint_ip and packet4.ip.dst == peer_ip):
                return True
            continue
        except IndexError:
            return False
    return False


def check_for_tunnel_packets(packets: list):
    return any(ifindex_to_name(packet.sll.ifindex) == "nordlynx" for packet in packets)
